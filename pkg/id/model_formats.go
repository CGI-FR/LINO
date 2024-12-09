package id

type idcolumnformat struct {
	imp string
	exp string
}

func NewIngressColumnFormat(imp string, exp string) IngressColumnFormat {
	return idcolumnformat{
		imp: imp,
		exp: exp,
	}
}

func (f idcolumnformat) Import() string { return f.imp }
func (f idcolumnformat) Export() string { return f.exp }

type idcolumnformatlist struct {
	formats map[string]IngressColumnFormat
}

func NewIngressColumnFormatList(formats map[string]IngressColumnFormat) IngressColumnFormatList {
	return idcolumnformatlist{
		formats: formats,
	}
}

func (l idcolumnformatlist) Len() uint                             { return uint(len(l.formats)) }
func (l idcolumnformatlist) Get(column string) IngressColumnFormat { return l.formats[column] }
