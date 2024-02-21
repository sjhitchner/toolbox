package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Flags().StringP("output", "o", "", "Output file")
	rootCmd.Flags().StringP("delimiter", "d", ",", "CSV delimiter")
	rootCmd.Flags().IntSliceP("fields", "f", []int{1}, "Field index")
}

var rootCmd = &cobra.Command{
	Use:   "csv-cut [file]",
	Short: "Cut CSV file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		delimiter, err := cmd.Flags().GetString("delimiter")
		if err != nil {
			return err
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}

		fields, err := cmd.Flags().GetIntSlice("fields")
		if err != nil {
			return err
		}

		filename := cmd.Flags().Arg(0)
		if filename == "" {
			// TODO determine is a pipe stdin
			return fmt.Errorf("filename empty")
		}

		in, err := getReader(filename)
		if err != nil {
			return err //log.Fatalf("Error opening CSV file: %s", err)
		}
		defer in.Close()

		out, err := getWriter(output)
		if err != nil {
			return err
		}
		defer out.Close()

		return processCSV(in, out, fields, delimiter)
	},
}

func processCSV(in io.Reader, out io.Writer, fields []int, delimiter string) error {

	// Create a new CSV reader
	reader := csv.NewReader(in)
	reader.Comma = getDelimiter(delimiter)

	writer := csv.NewWriter(out)
	writer.Comma = reader.Comma
	defer writer.Flush()

	output := make([]string, len(fields))

	// Read and process each record
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		for idx, field := range fields {

			if field > len(row) {
				return fmt.Errorf("Row doesn't contain enough fields %v", row)
			}

			output[idx] = row[field-1]
		}

		writer.Write(output)
	}

	return nil
}

func getDelimiter(delimiter string) rune {
	switch delimiter {
	case "space":
		return ' '
	case " ":
		return ' '
	case "tab":
		return '\t'
	case "\t":
		return '\t'
	case ",":
		return ','
	default:
		return rune(delimiter[0])
	}
}

func getReader(input string) (io.ReadCloser, error) {
	if input == "" {
		return os.Stdin, nil
	}

	f, err := os.Open(input)
	if err != nil {
		return nil, err //log.Fatalf("Error opening CSV file: %s", err)
	}
	return f, nil
}

func getWriter(output string) (io.WriteCloser, error) {
	if output == "" {
		return os.Stdout, nil
	}

	f, err := os.Create(output)
	if err != nil {
		return nil, err //log.Fatalf("Error opening CSV file: %s", err)
	}
	return f, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
