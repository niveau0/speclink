package sales

import (
	"iter"

	"example.com/erp/pkg/permtext"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
)

// StreamQuoteOverviews reads the quotation overview lazily, row by row.
//
// The result is a sequence and nothing beside it. A stream decides nothing at
// the moment it is handed over — not even whether the caller may read — so an
// error returned then could only ever be nil, and a caller who checked it would
// believe it had checked. The sequence carries both halves of the answer.
type StreamQuoteOverviews func(subject auth.Subject) iter.Seq2[QuoteOverview, error]

// PermStreamQuoteOverviews guards the streaming read.
var PermStreamQuoteOverviews = permission.Declare[StreamQuoteOverviews](
	"sales.quote.stream",
	permtext.Name("sales.quote.stream", "Stream quotes"),
	permtext.Desc("sales.quote.stream", "Holders may read the quotation list as a stream."),
)

// NewStreamQuoteOverviews builds the streaming read over the read model.
func NewStreamQuoteOverviews(all QuoteOverviewLister) StreamQuoteOverviews {
	return func(subject auth.Subject) iter.Seq2[QuoteOverview, error] {
		return func(yield func(QuoteOverview, error) bool) {
			if err := subject.Audit(PermStreamQuoteOverviews); err != nil {
				yield(QuoteOverview{}, err)
				return
			}

			for _, o := range all.All() {
				if !yield(o, nil) {
					return
				}
			}
		}
	}
}
