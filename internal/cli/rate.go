package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/store"
)

func newRateCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "rate <id> <+1|0|-1>",
		Short: "Rate a resource (+1, 0, or -1)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			r, err := s.GetResource(args[0])
			if err != nil {
				return err
			}
			rating, err := parseRating(args[1])
			if err != nil {
				return err
			}
			r.Rating = rating
			if err := s.SaveResource(r); err != nil {
				return err
			}
			fmt.Printf("Rated %q: %s\n", r.Title, formatRating(rating))
			return nil
		},
	}
}

func parseRating(s string) (*int, error) {
	var v int
	switch s {
	case "+1", "1":
		v = 1
	case "0":
		v = 0
	case "-1":
		v = -1
	default:
		return nil, fmt.Errorf("rating must be +1, 0, or -1 (got %q)", s)
	}
	return &v, nil
}
