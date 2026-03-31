package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/store"
)

var dataDir string

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "qlog",
		Short:   "questlog — your personal learning tracker",
		Version: Version,
		Long: `questlog (qlog) helps you capture ideas, track learning resources,
and focus your study sessions. Think of it as your RPG quest log for knowledge.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	defaultDataDir := filepath.Join(os.Getenv("HOME"), ".questlog")
	root.PersistentFlags().StringVar(&dataDir, "data-dir", defaultDataDir, "questlog data directory")

	// Store is initialised once in PersistentPreRunE and passed into each command via closure.
	// Commands that need the store use mustStore() which panics if not called after init.
	var s *store.FSStore

	mustStore := func() *store.FSStore {
		if s == nil {
			panic("store not initialised")
		}
		return s
	}

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		s, err = store.New(dataDir)
		if err != nil {
			return fmt.Errorf("failed to open questlog at %s: %w", dataDir, err)
		}
		return nil
	}

	root.AddCommand(
		newVersionCmd(),
		newAddCmd(mustStore),
		newInboxCmd(mustStore),
		newTrackCmd(mustStore),
		newListCmd(mustStore),
		newDoneCmd(mustStore),
		newProgressCmd(mustStore),
		newClassifyCmd(mustStore),
		newFocusCmd(mustStore),
		newGuideCmd(mustStore),
		newSearchCmd(mustStore),
		newStatusCmd(mustStore),
		newIndexCmd(mustStore),
	)

	return root
}
