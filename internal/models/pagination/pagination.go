package pagination

type Option func(opts *Options)

type Options struct {
	limit  int
	offset int // для простоты; на больших таблицах лучше Keyset-пагинацию
	sort   []SortField
}

func NewOptions(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func (p *Options) Limit() int {
	return p.limit
}

func (p *Options) Offset() int {
	return p.offset
}

func (p *Options) OrderBy() []SortField {
	return p.sort
}

type SortField struct {
	Name string
	Desc bool
}

var (
	ASC  = false
	DESC = true
)

func OrderBy(name string, desc bool) SortField {
	return SortField{Name: name, Desc: desc}
}

func WithLimit(limit int) Option {
	return func(opts *Options) { opts.limit = limit }
}

func WithOffset(offset int) Option {
	return func(opts *Options) { opts.offset = offset }
}

func WithSortFields(fields ...SortField) Option {
	return func(opts *Options) { opts.sort = fields }
}
