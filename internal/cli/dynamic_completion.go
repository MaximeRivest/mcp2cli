package cli

import (
	"strings"
	"time"

	"github.com/maximerivest/mcptocli/internal/cache"
	"github.com/maximerivest/mcptocli/internal/config"
	"github.com/maximerivest/mcptocli/internal/naming"
	"github.com/maximerivest/mcptocli/internal/schema/inspect"
	"github.com/maximerivest/mcptocli/internal/serverref"
	"github.com/spf13/cobra"
)

const completionCacheTTL = 24 * time.Hour

func serverNameCompletions(state *State, toComplete string) ([]string, cobra.ShellCompDirective) {
	repo, err := state.Repo()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	servers, err := repo.ListServers()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	matches := make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.HasPrefix(server.Name, strings.ToLower(toComplete)) {
			matches = append(matches, server.Name)
		}
	}
	return matches, cobra.ShellCompDirectiveNoFileComp
}

func metadataForCompletion(state *State, server *config.Server) *cache.Metadata {
	store, err := state.MetadataStore()
	if err != nil || store == nil {
		return nil
	}
	metadata, err := store.LoadFresh(server, completionCacheTTL)
	if err != nil {
		return nil
	}
	return metadata
}

func resolveCompletionServer(state *State, explicitName string) *config.Server {
	repo, err := state.Repo()
	if err != nil {
		return nil
	}
	bound, err := state.BoundServer()
	if err == nil && bound != nil {
		resolved, err := serverref.Resolve(repo, bound, serverref.Options{})
		if err == nil {
			return resolved.Server
		}
	}
	if explicitName == "" {
		return nil
	}
	resolved, err := serverref.Resolve(repo, nil, serverref.Options{ExplicitName: explicitName})
	if err != nil {
		return nil
	}
	return resolved.Server
}

func toolNameCompletions(state *State, server *config.Server, toComplete string) []string {
	metadata := metadataForCompletion(state, server)
	if metadata == nil {
		return nil
	}
	matches := []string{}
	for _, tool := range metadata.Tools {
		name := naming.ToKebabCase(tool.Name)
		if strings.HasPrefix(name, strings.ToLower(toComplete)) {
			matches = append(matches, name)
		}
	}
	return matches
}

func resourceNameCompletions(state *State, server *config.Server, toComplete string) []string {
	metadata := metadataForCompletion(state, server)
	if metadata == nil {
		return nil
	}
	matches := []string{}
	for _, resource := range metadata.Resources {
		name := resourceDisplayName(resource)
		if strings.HasPrefix(name, strings.ToLower(toComplete)) {
			matches = append(matches, name)
		}
	}
	return matches
}

func promptNameCompletions(state *State, server *config.Server, toComplete string) []string {
	metadata := metadataForCompletion(state, server)
	if metadata == nil {
		return nil
	}
	matches := []string{}
	for _, prompt := range metadata.Prompts {
		name := promptDisplayName(prompt)
		if strings.HasPrefix(name, strings.ToLower(toComplete)) {
			matches = append(matches, name)
		}
	}
	return matches
}

func toolFlagCompletions(state *State, server *config.Server, toolName, toComplete string) []string {
	metadata := metadataForCompletion(state, server)
	if metadata == nil {
		return nil
	}
	for _, tool := range metadata.Tools {
		if naming.ToKebabCase(tool.Name) != naming.ToKebabCase(toolName) {
			continue
		}
		spec, err := inspect.InspectTool(tool)
		if err != nil {
			return nil
		}
		matches := []string{}
		for _, arg := range spec.Arguments {
			flag := "--" + arg.CLIName
			if strings.HasPrefix(flag, toComplete) {
				matches = append(matches, flag)
			}
			if arg.Type == "boolean" {
				noFlag := "--no-" + arg.CLIName
				if strings.HasPrefix(noFlag, toComplete) {
					matches = append(matches, noFlag)
				}
			}
		}
		return matches
	}
	return nil
}

func promptFlagCompletions(state *State, server *config.Server, promptName, toComplete string) []string {
	metadata := metadataForCompletion(state, server)
	if metadata == nil {
		return nil
	}
	for _, prompt := range metadata.Prompts {
		if promptDisplayName(prompt) != naming.ToKebabCase(promptName) {
			continue
		}
		spec := inspect.InspectPrompt(prompt)
		matches := []string{}
		for _, arg := range spec.Arguments {
			flag := "--" + arg.CLIName
			if strings.HasPrefix(flag, toComplete) {
				matches = append(matches, flag)
			}
		}
		return matches
	}
	return nil
}

func metadataCompletionDirective() cobra.ShellCompDirective {
	return cobra.ShellCompDirectiveNoFileComp
}

func parseCompletionPositionals(tokens []string) []string {
	positionals := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token == "-o" {
			i++
			continue
		}
		if !strings.HasPrefix(token, "-") {
			positionals = append(positionals, token)
			continue
		}
		if strings.HasPrefix(token, "--") {
			flag := strings.TrimPrefix(token, "--")
			if before, _, ok := strings.Cut(flag, "="); ok {
				flag = before
			}
			if isKnownToolFlag(flag) {
				if !strings.Contains(token, "=") {
					i++
				}
				continue
			}
		}
	}
	return positionals
}

func metadataToolsCompletion(state *State, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if state.Options.Invocation.IsExposedCommand() {
		server := resolveCompletionServer(state, "")
		if len(args) == 0 {
			return toolNameCompletions(state, server, toComplete), metadataCompletionDirective()
		}
		return nil, metadataCompletionDirective()
	}
	if command, _ := cmd.Flags().GetString("command"); command != "" {
		return nil, metadataCompletionDirective()
	}
	if urlValue, _ := cmd.Flags().GetString("url"); urlValue != "" {
		return nil, metadataCompletionDirective()
	}
	if len(args) == 0 {
		return serverNameCompletions(state, toComplete)
	}
	server := resolveCompletionServer(state, args[0])
	if len(args) == 1 {
		return toolNameCompletions(state, server, toComplete), metadataCompletionDirective()
	}
	return nil, metadataCompletionDirective()
}

func metadataResourcesCompletion(state *State, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if state.Options.Invocation.IsExposedCommand() {
		server := resolveCompletionServer(state, "")
		if len(args) == 0 {
			return resourceNameCompletions(state, server, toComplete), metadataCompletionDirective()
		}
		return nil, metadataCompletionDirective()
	}
	if command, _ := cmd.Flags().GetString("command"); command != "" {
		return nil, metadataCompletionDirective()
	}
	if urlValue, _ := cmd.Flags().GetString("url"); urlValue != "" {
		return nil, metadataCompletionDirective()
	}
	if len(args) == 0 {
		return serverNameCompletions(state, toComplete)
	}
	server := resolveCompletionServer(state, args[0])
	if len(args) == 1 {
		return resourceNameCompletions(state, server, toComplete), metadataCompletionDirective()
	}
	return nil, metadataCompletionDirective()
}

func metadataPromptsCompletion(state *State, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if state.Options.Invocation.IsExposedCommand() {
		server := resolveCompletionServer(state, "")
		if len(args) == 0 {
			return promptNameCompletions(state, server, toComplete), metadataCompletionDirective()
		}
		return nil, metadataCompletionDirective()
	}
	if command, _ := cmd.Flags().GetString("command"); command != "" {
		return nil, metadataCompletionDirective()
	}
	if urlValue, _ := cmd.Flags().GetString("url"); urlValue != "" {
		return nil, metadataCompletionDirective()
	}
	if len(args) == 0 {
		return serverNameCompletions(state, toComplete)
	}
	server := resolveCompletionServer(state, args[0])
	if len(args) == 1 {
		return promptNameCompletions(state, server, toComplete), metadataCompletionDirective()
	}
	return nil, metadataCompletionDirective()
}

func toolCommandCompletion(state *State, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	positionals := parseCompletionPositionals(args)
	boundOrDirect := state.Options.Invocation.IsExposedCommand() || hasDirectConnectionFlags(args)
	if !boundOrDirect {
		if len(positionals) == 0 && !strings.HasPrefix(toComplete, "-") {
			return serverNameCompletions(state, toComplete)
		}
		if len(positionals) == 1 && !strings.HasPrefix(toComplete, "-") {
			server := resolveCompletionServer(state, positionals[0])
			return toolNameCompletions(state, server, toComplete), metadataCompletionDirective()
		}
		if len(positionals) >= 2 && strings.HasPrefix(toComplete, "-") {
			server := resolveCompletionServer(state, positionals[0])
			return toolFlagCompletions(state, server, positionals[1], toComplete), metadataCompletionDirective()
		}
		return nil, metadataCompletionDirective()
	}
	server := resolveCompletionServer(state, "")
	if len(positionals) == 0 && !strings.HasPrefix(toComplete, "-") {
		return toolNameCompletions(state, server, toComplete), metadataCompletionDirective()
	}
	if len(positionals) >= 1 && strings.HasPrefix(toComplete, "-") {
		return toolFlagCompletions(state, server, positionals[0], toComplete), metadataCompletionDirective()
	}
	return nil, metadataCompletionDirective()
}

func promptCommandCompletion(state *State, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	positionals := parseCompletionPositionals(args)
	boundOrDirect := state.Options.Invocation.IsExposedCommand() || hasDirectConnectionFlags(args)
	if !boundOrDirect {
		if len(positionals) == 0 && !strings.HasPrefix(toComplete, "-") {
			return serverNameCompletions(state, toComplete)
		}
		if len(positionals) == 1 && !strings.HasPrefix(toComplete, "-") {
			server := resolveCompletionServer(state, positionals[0])
			return promptNameCompletions(state, server, toComplete), metadataCompletionDirective()
		}
		if len(positionals) >= 2 && strings.HasPrefix(toComplete, "-") {
			server := resolveCompletionServer(state, positionals[0])
			return promptFlagCompletions(state, server, positionals[1], toComplete), metadataCompletionDirective()
		}
		return nil, metadataCompletionDirective()
	}
	server := resolveCompletionServer(state, "")
	if len(positionals) == 0 && !strings.HasPrefix(toComplete, "-") {
		return promptNameCompletions(state, server, toComplete), metadataCompletionDirective()
	}
	if len(positionals) >= 1 && strings.HasPrefix(toComplete, "-") {
		return promptFlagCompletions(state, server, positionals[0], toComplete), metadataCompletionDirective()
	}
	return nil, metadataCompletionDirective()
}

func hasDirectConnectionFlags(tokens []string) bool {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token == "--command" || token == "--url" {
			return true
		}
		if strings.HasPrefix(token, "--command=") || strings.HasPrefix(token, "--url=") {
			return true
		}
	}
	return false
}
