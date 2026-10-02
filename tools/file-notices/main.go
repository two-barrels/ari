// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

// file-notices checks or adds per-file attribution relative to an upstream commit.
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func git(root string, args ...string) []byte {
	c := exec.Command("git", append([]string{"-C", root}, args...)...)
	b, err := c.Output()
	if err != nil {
		panic(err)
	}
	return b
}

func main() {
	root := flag.String("repo", ".", "repository to inspect")
	baseline := flag.String("baseline", "90f525536787cb75fdf89b1e57f6081f4b93d908", "pinned upstream comparison commit")
	apply := flag.Bool("apply", false, "add missing notices; default checks only")
	flag.Parse()
	// Include working changes and added files, but never .git or ignored artifacts.
	paths := map[string]bool{}
	for _, p := range strings.Split(string(git(*root, "diff", "--name-only", "--diff-filter=ACMRT", *baseline)), "\n") {
		if p != "" {
			paths[p] = true
		}
	}
	for _, p := range strings.Split(string(git(*root, "ls-files", "--others", "--exclude-standard")), "\n") {
		if p != "" {
			paths[p] = true
		}
	}
	names := make([]string, 0, len(paths))
	for p := range paths {
		names = append(names, p)
	}
	sort.Strings(names)
	failures := 0
	for _, p := range names {
		if p == "LICENSE" {
			continue
		} // Preserve upstream license verbatim.
		name := filepath.Join(*root, p)
		b, err := os.ReadFile(name)
		if err != nil {
			panic(err)
		}
		exists := exec.Command("git", "-C", *root, "cat-file", "-e", *baseline+":"+p).Run() == nil
		notice := "Created by two-barrels in 2026 for ARI v6 modernization."
		if exists {
			notice = "Modified by two-barrels in 2026 for ARI v6 modernization."
		}
		ext := filepath.Ext(p)
		sidecar := ext == ".sum" || strings.HasSuffix(p, "/json/events-23.json")
		if strings.HasSuffix(p, ".notice") || p == "NOTICE" {
			if bytes.Contains(b, []byte("two-barrels")) {
				continue
			}
		}
		if sidecar {
			nb, e := os.ReadFile(name + ".notice")
			if e == nil && bytes.Contains(nb, []byte(notice)) {
				continue
			}
			if !*apply {
				fmt.Println("missing companion notice:", p)
				failures++
				continue
			}
			text := "Created by two-barrels in 2026 as a file attribution notice.\n" + notice + "\nSPDX-License-Identifier: Apache-2.0\n\nThis notice accompanies " + p + "; its machine-readable content is preserved.\nThis notice does not relicense third-party content.\n"
			if ext == ".json" {
				text = "Added by two-barrels in 2026 for ARI v6 modernization.\nCopied verbatim from Asterisk events.json at\n97d55b3306ca4aa26c0136c67b79470ca4b2b785.\nOriginal Digium copyright and author notices remain in the JSON file.\n\n" + text
			}
			if err := os.WriteFile(name+".notice", []byte(text), 0644); err != nil {
				panic(err)
			}
			continue
		}
		if bytes.Contains(b, []byte(notice)) {
			continue
		}
		if !*apply {
			fmt.Println("missing notice:", p)
			failures++
			continue
		}
		var out []byte
		switch ext {
		case ".go", ".mod", ".tmpl":
			out = append([]byte("// "+notice+"\n// SPDX-License-Identifier: Apache-2.0\n\n"), b...)
		case ".md":
			out = append([]byte("<!-- "+notice+" SPDX-License-Identifier: Apache-2.0 -->\n\n"), b...)
		case ".yaml", ".yml", ".gitignore":
			out = append([]byte("# "+notice+"\n# SPDX-License-Identifier: Apache-2.0\n\n"), b...)
		case ".json":
			if !json.Valid(b) {
				panic("invalid JSON: " + p)
			}
			out = append([]byte("{\n  \"_notice\": "+fmt.Sprintf("%q", notice+" SPDX-License-Identifier: Apache-2.0")+","), bytes.TrimPrefix(bytes.TrimSpace(b), []byte("{"))...)
			out = append(out, '\n')
		case ".csv":
			rows, err := csv.NewReader(bytes.NewReader(b)).ReadAll()
			if err != nil {
				panic(err)
			}
			for i := range rows {
				v := notice + " SPDX-License-Identifier: Apache-2.0"
				if i == 0 {
					v = "file_notice"
				}
				rows[i] = append(rows[i], v)
			}
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			w.WriteAll(rows)
			if w.Error() != nil {
				panic(w.Error())
			}
			out = buf.Bytes()
		default:
			if p == "Makefile" {
				out = append([]byte("# "+notice+"\n# SPDX-License-Identifier: Apache-2.0\n\n"), b...)
			} else {
				panic("unsupported file: " + p)
			}
		}
		if err := os.WriteFile(name, out, 0644); err != nil {
			panic(err)
		}
	}
	if failures > 0 {
		os.Exit(1)
	}
	fmt.Printf("File notices OK: %d files inspected against %s\n", len(names), *baseline)
}
