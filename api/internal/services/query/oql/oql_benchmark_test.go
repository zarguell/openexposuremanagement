package oql

import (
	"testing"
)

// BenchmarkTokenize_SimpleQuery benchmarks tokenizing a simple query
func BenchmarkTokenize_SimpleQuery(b *testing.B) {
	input := "is_active = true limit 10"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Tokenize(input)
	}
}

// BenchmarkTokenize_ComplexQuery benchmarks tokenizing a complex query
func BenchmarkTokenize_ComplexQuery(b *testing.B) {
	input := `is_active = true AND (software.vendor = "Microsoft" OR software.vendor = "Oracle") AND NOT findings.severity = "low" sort canonical_name desc limit 20 offset 40`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Tokenize(input)
	}
}

// BenchmarkTokenize_DeepDotWalking benchmarks tokenizing deep dot-walking
func BenchmarkTokenize_DeepDotWalking(b *testing.B) {
	input := `findings.definition.cve = "CVE-2021-44228" AND software.cpe_23 like "cpe:2.3:a:apache:%"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Tokenize(input)
	}
}

// BenchmarkParse_SimpleQuery benchmarks parsing a simple query
func BenchmarkParse_SimpleQuery(b *testing.B) {
	input := "is_active = true limit 10"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(input)
	}
}

// BenchmarkParse_ComplexQuery benchmarks parsing a complex query
func BenchmarkParse_ComplexQuery(b *testing.B) {
	input := `is_active = true AND (findings.severity = "critical" OR findings.epss_score > 0.9) sort last_seen_at desc limit 20`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(input)
	}
}

// BenchmarkParse_MultipleConditions benchmarks parsing with multiple AND/OR conditions
func BenchmarkParse_MultipleConditions(b *testing.B) {
	input := `is_active = true AND software.vendor = "Microsoft" AND findings.severity = "critical" AND findings.epss_score > 0.9`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(input)
	}
}

// BenchmarkParse_NestedExpressions benchmarks parsing nested logical expressions
func BenchmarkParse_NestedExpressions(b *testing.B) {
	input := `(is_active = true AND software.vendor = "Microsoft") OR (is_active = false AND findings.severity = "critical")`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(input)
	}
}

// BenchmarkParseOQL_WithFilters benchmarks ParseOQL with filters
func BenchmarkParseOQL_WithFilters(b *testing.B) {
	input := `is_active = true AND software.vendor = "Microsoft"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseOQL(input)
	}
}

// BenchmarkParseOQL_ComplexFilters benchmarks ParseOQL with complex filters
func BenchmarkParseOQL_ComplexFilters(b *testing.B) {
	input := `is_active = true AND (software.vendor = "Microsoft" OR software.vendor = "Oracle") AND findings.severity = "critical"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseOQL(input)
	}
}

// BenchmarkParseOQL_FullPipeline benchmarks the full pipeline (tokenize + parse + translate)
func BenchmarkParseOQL_FullPipeline(b *testing.B) {
	input := `is_active = true AND software.vendor = "Microsoft" sort canonical_name limit 10`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseOQL(input)
	}
}

// BenchmarkParseOQL_Simple benchmarks ParseOQL with simple query
func BenchmarkParseOQL_Simple(b *testing.B) {
	input := "is_active = true"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseOQL(input)
	}
}

// BenchmarkParseOQL_WithSort benchmarks ParseOQL with sort clause
func BenchmarkParseOQL_WithSort(b *testing.B) {
	input := `is_active = true sort canonical_name desc limit 10`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseOQL(input)
	}
}

// BenchmarkParseOQL_WithLimit benchmarks ParseOQL with limit/offset
func BenchmarkParseOQL_WithLimit(b *testing.B) {
	input := `is_active = true limit 50 offset 100`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseOQL(input)
	}
}

// BenchmarkParseOQL_InOperator benchmarks ParseOQL with IN operator
func BenchmarkParseOQL_InOperator(b *testing.B) {
	input := `findings.severity in ("critical", "high", "medium")`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseOQL(input)
	}
}

// BenchmarkParseOQL_LongQuery benchmarks ParseOQL with long realistic query
func BenchmarkParseOQL_LongQuery(b *testing.B) {
	input := `is_active = true AND (findings.severity = "critical" OR findings.epss_score > 0.9) AND NOT software.vendor = "CrowdStrike" sort last_seen_at desc limit 20`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseOQL(input)
	}
}

// BenchmarkTokenize_LongInput benchmarks tokenizing progressively longer inputs
func BenchmarkTokenize_Short(b *testing.B) {
	input := "is_active = true"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tokenize(input)
	}
}

func BenchmarkTokenize_Medium(b *testing.B) {
	input := `is_active = true AND software.vendor = "Microsoft" AND findings.severity = "critical"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tokenize(input)
	}
}

func BenchmarkTokenize_Long(b *testing.B) {
	input := `is_active = true AND (software.vendor = "Microsoft" OR software.vendor = "Oracle") AND (findings.severity = "critical" OR findings.epss_score > 0.9) AND NOT software.vendor = "CrowdStrike" sort canonical_name desc limit 20`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tokenize(input)
	}
}
