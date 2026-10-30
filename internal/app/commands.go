package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// parseFlags parses args into fs. Help requests (-h, -help, --help) print the
// flag usage to stdout and report help=true; parse errors are written to
// stderr. Callers return 0 for help and 2 for any other error.
func parseFlags(fs *flag.FlagSet, usageLine string, args []string) (help bool, err error) {
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.Usage = func() {
		fmt.Fprintf(&buf, "Usage: %s\n\nFlags:\n", usageLine)
		fs.PrintDefaults()
	}
	err = fs.Parse(args)
	switch {
	case err == nil:
		return false, nil
	case errors.Is(err, flag.ErrHelp):
		_, _ = buf.WriteTo(os.Stdout)
		return true, nil
	default:
		_, _ = buf.WriteTo(os.Stderr)
		return false, err
	}
}

func runScan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	trainPath := fs.String("train", "", "path to a train JSONL file")
	evalPath := fs.String("eval", "", "path to an eval JSONL file")
	semanticThreshold := fs.Float64("semantic-threshold", 0.6, "token-set similarity threshold for semantic duplicates")
	maxSamples := fs.Int("max-samples", 5, "maximum number of example pairs to print for each issue")
	strict := fs.Bool("strict", false, "exit with code 1 if any issue is found")
	jsonOutput := fs.Bool("json", false, "emit machine-readable JSON")

	if help, err := parseFlags(fs, "datasetlint scan -train FILE -eval FILE [-semantic-threshold F] [-max-samples N] [-strict] [-json]", args); err != nil {
		return 2
	} else if help {
		return 0
	}
	if *trainPath == "" || *evalPath == "" {
		fmt.Fprintln(os.Stderr, "both -train and -eval are required")
		return 2
	}

	train, err := loadDataset(*trainPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	eval, err := loadDataset(*evalPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	report := scanDatasetsWithOptions(train, eval, *semanticThreshold, *maxSamples)

	if *jsonOutput {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(string(body))
		if *strict && hasIssues(report) {
			return 1
		}
		return 0
	}

	printReport(report)
	if *strict && hasIssues(report) {
		return 1
	}
	return 0
}

func printReport(report Report) {
	fmt.Printf("Train rows: %d\n", report.Train.Rows)
	fmt.Printf("Eval rows: %d\n", report.Eval.Rows)
	fmt.Printf("Train issues: missing id=%d empty input=%d empty output=%d duplicate inputs=%d near duplicates=%d label conflicts=%d\n",
		report.Train.MissingID, report.Train.EmptyInput, report.Train.EmptyOutput, report.Train.DuplicateInputs, report.Train.NearDuplicatePairs, report.Train.LabelConflicts)
	fmt.Printf("Eval issues: missing id=%d empty input=%d empty output=%d duplicate inputs=%d near duplicates=%d label conflicts=%d\n",
		report.Eval.MissingID, report.Eval.EmptyInput, report.Eval.EmptyOutput, report.Eval.DuplicateInputs, report.Eval.NearDuplicatePairs, report.Eval.LabelConflicts)
	fmt.Printf("Cross-split overlap: %d\n", report.OverlapCount)
	fmt.Printf("Cross-split semantic overlap: %d\n", report.SemanticOverlapCount)
	fmt.Printf("Train lengths avg/p95: %.1f/%d input tokens, %.1f/%d output tokens\n",
		report.Train.LengthStats.AvgInputTokens, report.Train.LengthStats.P95InputTokens, report.Train.LengthStats.AvgOutputTokens, report.Train.LengthStats.P95OutputTokens)
	fmt.Printf("Eval lengths avg/p95: %.1f/%d input tokens, %.1f/%d output tokens\n",
		report.Eval.LengthStats.AvgInputTokens, report.Eval.LengthStats.P95InputTokens, report.Eval.LengthStats.AvgOutputTokens, report.Eval.LengthStats.P95OutputTokens)
	if len(report.Train.LabelCounts) > 0 {
		fmt.Printf("Train labels: %v\n", report.Train.LabelCounts)
	}
	if len(report.Eval.LabelCounts) > 0 {
		fmt.Printf("Eval labels: %v\n", report.Eval.LabelCounts)
	}
	if len(report.OverlapExamples) > 0 {
		fmt.Printf("Overlap examples: %s\n", strings.Join(report.OverlapExamples, "; "))
	}
	if len(report.SemanticOverlapSamples) > 0 {
		fmt.Printf("Semantic overlap example: %s <-> %s (%.2f)\n",
			report.SemanticOverlapSamples[0].LeftID, report.SemanticOverlapSamples[0].RightID, report.SemanticOverlapSamples[0].Similarity)
	}
}
