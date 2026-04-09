package cli

import (
	"config-analyzer/internal/infrastructure/config"
	"config-analyzer/internal/service/analyzer"

	"github.com/spf13/cobra"
)

var (
	format string
	silent bool
	stdin  bool
)

func init() {
	analyzeCmd.SilenceErrors = true
	analyzeCmd.SilenceUsage = true
	analyzeCmd.Flags().BoolVarP(&silent, "silent", "s", false, "Не выходить с ошибкой при проблемах конфига")
	analyzeCmd.Flags().BoolVar(&stdin, "stdin", false, "Читать конфигурацию из стандартного ввода")
	analyzeCmd.Flags().StringVarP(&format, "format", "f", "", "Формат конфига: JSON или YAML (опционально)")
}

func Execute() error {
	if err := analyzeCmd.Execute(); err != nil {
		return err
	}
	return nil
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Анализатор конфигураций YAML/JSON",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(args, stdin, format)
		if err != nil {
			return err
		}

		return analyzer.AnalyzeAndPrint(cfg, silent)
	},
}
