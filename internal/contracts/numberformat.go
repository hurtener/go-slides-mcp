package contracts

// NumberFormat is a deterministic number / currency / percent / locale
// format (R14.13, D-121). Mirror of pptx-go's scene.NumberFormat: the
// caller supplies it (e.g. a soul's number token or a per-Stat override),
// the engine only formats with it — it never decides the format itself
// (D-026). A Stat's Number is rendered through FormatNumber(Number, Format).
//
// The zero value formats with no grouping, no decimals, and no affixes
// (e.g. 4000 -> "4000"). A en-US currency sets GroupSep "," and
// CurrencySymbol "$" (4000 -> "$4,000"); a de-DE locale sets GroupSep "."
// and DecimalSep "," (4000 -> "4.000"). Layout order is Prefix, sign,
// [symbol if !SymbolAfter], body, [%], [symbol if SymbolAfter], Suffix.
type NumberFormat struct {
	// Decimals is the fixed number of decimal places (0 = integer). Under
	// Compact notation, 0 is treated as 1 (so 1200000 -> "1.2M", not "1M").
	Decimals int `json:"decimals,omitempty"`
	// GroupSep is the thousands separator (e.g. "," or "."); "" = no
	// grouping.
	GroupSep string `json:"groupSep,omitempty"`
	// DecimalSep is the decimal point (e.g. "," for de-DE); "" defaults
	// to ".".
	DecimalSep string `json:"decimalSep,omitempty"`
	// CurrencySymbol is prepended (or appended, see SymbolAfter); "" =
	// none.
	CurrencySymbol string `json:"currencySymbol,omitempty"`
	// SymbolAfter places the currency symbol after the number (e.g.
	// "4.000 €") instead of before it (e.g. "$4,000").
	SymbolAfter bool `json:"symbolAfter,omitempty"`
	// Percent multiplies the value by 100 and appends "%" (e.g.
	// 0.92 -> "92%").
	Percent bool `json:"percent,omitempty"`
	// Compact renders large magnitudes as K / M / B / T (e.g.
	// 1200000 -> "1.2M").
	Compact bool `json:"compact,omitempty"`
	// CompactThreshold is the magnitude at/above which Compact applies;
	// 0 = 1000. Ignored when Compact is false.
	CompactThreshold float64 `json:"compactThreshold,omitempty"`
	// Prefix is an arbitrary string prepended before the sign/symbol/body
	// (e.g. "~" for an approximate value).
	Prefix string `json:"prefix,omitempty"`
	// Suffix is an arbitrary string appended after the body/%/symbol
	// (e.g. "+" for "$4,000+").
	Suffix string `json:"suffix,omitempty"`
}
