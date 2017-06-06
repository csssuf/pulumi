// Licensed to Pulumi Corporation ("Pulumi") under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// Pulumi licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/pulumi/lumi/pkg/compiler/core"
	"github.com/pulumi/lumi/pkg/eval/heapstate"
	"github.com/pulumi/lumi/pkg/graph"
	"github.com/pulumi/lumi/pkg/graph/dotconv"
	"github.com/pulumi/lumi/pkg/resource"
	"github.com/pulumi/lumi/pkg/tokens"
	"github.com/pulumi/lumi/pkg/util/cmdutil"
)

func newPackEvalCmd() *cobra.Command {
	var configEnv string
	var dotOutput bool
	var cmd = &cobra.Command{
		Use:   "eval [package] [-- [args]]",
		Short: "Evaluate a package and print the resulting objects",
		Long: "Evaluate a package and print the resulting objects\n" +
			"\n" +
			"A graph is a topologically sorted directed-acyclic-graph (DAG), representing a\n" +
			"collection of resources that may be used in a deployment operation like plan or apply.\n" +
			"This graph is produced by evaluating the contents of a blueprint package, and does not\n" +
			"actually perform any updates to the target environment.\n" +
			"\n" +
			"By default, a blueprint package is loaded from the current directory.  Optionally,\n" +
			"a path to a package elsewhere can be provided as the [package] argument.",
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, args []string) error {
			// If a configuration environment was requested, load it.
			var config resource.ConfigMap
			if configEnv != "" {
				envInfo, err := initEnvCmdName(tokens.QName(configEnv), args)
				if err != nil {
					return err
				}
				config = envInfo.Env.Config
			}

			// Perform the compilation and, if non-nil is returned, output the graph.
			if result := compile(cmd, args, config); result != nil && result.Heap != nil && result.Heap.G != nil {
				// Serialize that evaluation graph so that it's suitable for printing/serializing.
				if dotOutput {
					// Convert the output to a DOT file.
					if err := dotconv.Print(result.Heap.G, os.Stdout); err != nil {
						return errors.Errorf("failed to write DOT file to output: %v", err)
					}
				} else {
					// Just print a very basic, yet (hopefully) aesthetically pleasing, ascii-ization of the graph.
					shown := make(map[graph.Vertex]bool)
					for _, root := range result.Heap.G.Objs() {
						printVertex(root.ToObj(), shown, "")
					}
				}
			}
			return nil
		}),
	}

	cmd.PersistentFlags().StringVar(
		&configEnv, "config-env", "",
		"Apply configuration from the specified environment before evaluating the package")
	cmd.PersistentFlags().BoolVar(
		&dotOutput, "dot", false,
		"Output the graph as a DOT digraph (graph description language)")

	return cmd
}

// printVertex just pretty-prints a graph.  The output is not serializable, it's just for display purposes.
// IDEA: option to print properties.
// IDEA: full serializability, including a DOT file option.
func printVertex(v *heapstate.ObjectVertex, shown map[graph.Vertex]bool, indent string) {
	s := v.Obj().Type()
	if shown[v] {
		fmt.Printf("%v%v: <cycle...>\n", indent, s)
	} else {
		shown[v] = true // prevent cycles.
		fmt.Printf("%v%v:\n", indent, s)
		for _, out := range v.OutObjs() {
			printVertex(out.ToObj(), shown, indent+"    -> ")
		}
	}
}

// dashdashArgsToMap is a simple args parser that places incoming key/value pairs into a map.  These are then used
// during package compilation as inputs to the main entrypoint function.
// IDEA: this is fairly rudimentary; we eventually want to support arrays, maps, and complex types.
func dashdashArgsToMap(args []string) core.Args {
	mapped := make(core.Args)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Eat - or -- at the start.
		if arg[0] == '-' {
			arg = arg[1:]
			if arg[0] == '-' {
				arg = arg[1:]
			}
		}

		// Now find a k=v, and split the k/v part.
		if eq := strings.IndexByte(arg, '='); eq != -1 {
			// For --k=v, simply store v underneath k's entry.
			mapped[tokens.Name(arg[:eq])] = arg[eq+1:]
		} else {
			if i+1 < len(args) && args[i+1][0] != '-' {
				// If the next arg doesn't start with '-' (i.e., another flag) use its value.
				mapped[tokens.Name(arg)] = args[i+1]
				i++
			} else if arg[0:3] == "no-" {
				// For --no-k style args, strip off the no- prefix and store false underneath k.
				mapped[tokens.Name(arg[3:])] = false
			} else {
				// For all other --k args, assume this is a boolean flag, and set the value of k to true.
				mapped[tokens.Name(arg)] = true
			}
		}
	}

	return mapped
}
