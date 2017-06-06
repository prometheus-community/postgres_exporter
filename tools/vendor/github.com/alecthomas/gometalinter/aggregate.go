package main

import (
	"sort"
	"strings"
)

type (
	issueKey struct {
		path      string
		line, col int
		message   string
	}

	multiIssue struct {
		*Issue
		linterNames []string
	}
)

func aggregateIssues(issues chan *Issue) chan *Issue {
	out := make(chan *Issue, 1000000)
	issueMap := make(map[issueKey]*multiIssue)
	go func() {
		for issue := range issues {
			key := issueKey{
				path:    issue.Path,
				line:    issue.Line,
				col:     issue.Col,
				message: issue.Message,
			}
			if existing, ok := issueMap[key]; ok {
				existing.linterNames = append(existing.linterNames, issue.Linter.Name)
			} else {
				issueMap[key] = &multiIssue{
					Issue:       issue,
					linterNames: []string{issue.Linter.Name},
				}
			}
		}
		for _, multi := range issueMap {
			issue := multi.Issue
			sort.Strings(multi.linterNames)
			issue.Linter.Name = strings.Join(multi.linterNames, ", ")
			out <- issue
		}
		close(out)
	}()
	return out
}
