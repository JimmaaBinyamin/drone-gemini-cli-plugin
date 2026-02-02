package plugin

import (
	"fmt"
	"strings"
)

// ModelPricing contains pricing information for a model
type ModelPricing struct {
	Name                 string
	InputPriceShort      float64 // per 1M tokens (context <= 200K)
	InputPriceLong       float64 // per 1M tokens (context > 200K)
	OutputPriceShort     float64 // per 1M tokens
	OutputPriceLong      float64 // per 1M tokens
	LongContextThreshold int     // tokens, 0 means no long context pricing
}

// PricingTable contains pricing for all supported models
var PricingTable = map[string]ModelPricing{
	// Gemini 3.0 Series (Preview)
	"gemini-3-pro-preview": {
		Name:             "Gemini 3 Pro",
		InputPriceShort:  4.00,
		InputPriceLong:   4.00,
		OutputPriceShort: 12.00,
		OutputPriceLong:  12.00,
	},
	"gemini-3-flash-preview": {
		Name:             "Gemini 3 Flash",
		InputPriceShort:  0.50,
		InputPriceLong:   0.50,
		OutputPriceShort: 3.00,
		OutputPriceLong:  3.00,
	},

	// Gemini 2.5 Series (Production)
	"gemini-2.5-pro": {
		Name:                 "Gemini 2.5 Pro",
		InputPriceShort:      1.25,
		InputPriceLong:       2.50,
		OutputPriceShort:     10.00,
		OutputPriceLong:      15.00,
		LongContextThreshold: 200000,
	},
	"gemini-2.5-flash": {
		Name:             "Gemini 2.5 Flash",
		InputPriceShort:  0.30,
		InputPriceLong:   0.30,
		OutputPriceShort: 2.50,
		OutputPriceLong:  2.50,
	},
	"gemini-2.5-flash-lite": {
		Name:             "Gemini 2.5 Flash-Lite",
		InputPriceShort:  0.10,
		InputPriceLong:   0.10,
		OutputPriceShort: 0.40,
		OutputPriceLong:  0.40,
	},

	// Gemini 2.0 Series
	"gemini-2.0-flash": {
		Name:             "Gemini 2.0 Flash",
		InputPriceShort:  0.15,
		InputPriceLong:   0.15,
		OutputPriceShort: 0.60,
		OutputPriceLong:  0.60,
	},
	"gemini-2.0-flash-exp": {
		Name:             "Gemini 2.0 Flash (Exp)",
		InputPriceShort:  0.15,
		InputPriceLong:   0.15,
		OutputPriceShort: 0.60,
		OutputPriceLong:  0.60,
	},
	"gemini-2.0-flash-lite": {
		Name:             "Gemini 2.0 Flash-Lite",
		InputPriceShort:  0.075,
		InputPriceLong:   0.075,
		OutputPriceShort: 0.30,
		OutputPriceLong:  0.30,
	},

	// Gemini 1.5 Series (legacy)
	"gemini-1.5-pro": {
		Name:             "Gemini 1.5 Pro",
		InputPriceShort:  1.25,
		InputPriceLong:   1.25,
		OutputPriceShort: 5.00,
		OutputPriceLong:  5.00,
	},
	"gemini-1.5-flash": {
		Name:             "Gemini 1.5 Flash",
		InputPriceShort:  0.075,
		InputPriceLong:   0.075,
		OutputPriceShort: 0.30,
		OutputPriceLong:  0.30,
	},
}

// FormatStats formats CLI statistics as a readable string
func FormatStats(stats *CLIStats) string {
	if stats == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                    📊 执行统计                                ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")

	// Model statistics
	if len(stats.Models) > 0 {
		sb.WriteString("║  🤖 模型使用:                                                 ║\n")
		totalInputTokens := 0
		totalOutputTokens := 0
		totalThoughtsTokens := 0
		totalCost := 0.0

		for modelName, modelStats := range stats.Models {
			sb.WriteString(fmt.Sprintf("║    %-58s ║\n", modelName))
			sb.WriteString(fmt.Sprintf("║      请求: %d, 错误: %d, 延迟: %dms%s\n",
				modelStats.API.TotalRequests,
				modelStats.API.TotalErrors,
				modelStats.API.TotalLatencyMs,
				strings.Repeat(" ", 20)))

			tokens := modelStats.Tokens
			sb.WriteString(fmt.Sprintf("║      输入: %d, 输出: %d, 缓存: %d%s\n",
				tokens.Prompt,
				tokens.Candidates,
				tokens.Cached,
				strings.Repeat(" ", 20)))

			if tokens.Thoughts > 0 {
				sb.WriteString(fmt.Sprintf("║      🧠 思考 Tokens: %d%s\n",
					tokens.Thoughts,
					strings.Repeat(" ", 35)))
			}

			totalInputTokens += tokens.Prompt
			totalOutputTokens += tokens.Candidates
			totalThoughtsTokens += tokens.Thoughts

			// Calculate cost
			cost := calculateModelCost(modelName, tokens)
			totalCost += cost
		}

		sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
		sb.WriteString(fmt.Sprintf("║  总输入 Tokens: %-46d ║\n", totalInputTokens))
		sb.WriteString(fmt.Sprintf("║  总输出 Tokens: %-46d ║\n", totalOutputTokens))
		if totalThoughtsTokens > 0 {
			sb.WriteString(fmt.Sprintf("║  🧠 总思考 Tokens: %-43d ║\n", totalThoughtsTokens))
		}
		sb.WriteString(fmt.Sprintf("║  💵 预估成本: $%-47.6f ║\n", totalCost))
	}

	// Tool statistics
	if stats.Tools.TotalCalls > 0 {
		sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
		sb.WriteString("║  🔧 工具调用:                                                 ║\n")
		sb.WriteString(fmt.Sprintf("║    总调用: %d, 成功: %d, 失败: %d%s\n",
			stats.Tools.TotalCalls,
			stats.Tools.TotalSuccess,
			stats.Tools.TotalFail,
			strings.Repeat(" ", 25)))
		sb.WriteString(fmt.Sprintf("║    总耗时: %dms%s\n",
			stats.Tools.TotalDurationMs,
			strings.Repeat(" ", 43)))

		if len(stats.Tools.ByName) > 0 {
			sb.WriteString("║    工具详情:                                                  ║\n")
			for toolName, detail := range stats.Tools.ByName {
				sb.WriteString(fmt.Sprintf("║      - %s: %d次 (%dms)%s\n",
					toolName,
					detail.Count,
					detail.DurationMs,
					strings.Repeat(" ", 30)))
			}
		}
	}

	// File statistics
	if stats.Files.TotalLinesAdded > 0 || stats.Files.TotalLinesRemoved > 0 {
		sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
		sb.WriteString("║  📁 文件修改:                                                 ║\n")
		sb.WriteString(fmt.Sprintf("║    +%d 行添加, -%d 行删除%s\n",
			stats.Files.TotalLinesAdded,
			stats.Files.TotalLinesRemoved,
			strings.Repeat(" ", 32)))
	}

	sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// calculateModelCost calculates the cost for a model's token usage
func calculateModelCost(modelName string, tokens TokenStats) float64 {
	pricing, ok := PricingTable[modelName]
	if !ok {
		// Try partial match
		for key, p := range PricingTable {
			if strings.Contains(strings.ToLower(modelName), strings.ToLower(key)) {
				pricing = p
				ok = true
				break
			}
		}
	}

	if !ok {
		// Default pricing
		pricing = ModelPricing{
			InputPriceShort:  1.00,
			OutputPriceShort: 5.00,
		}
	}

	inputCost := float64(tokens.Prompt) / 1_000_000 * pricing.InputPriceShort
	outputCost := float64(tokens.Candidates) / 1_000_000 * pricing.OutputPriceShort
	thoughtsCost := float64(tokens.Thoughts) / 1_000_000 * pricing.OutputPriceShort

	return inputCost + outputCost + thoughtsCost
}

// FormatStatsSimple returns a one-line summary
func FormatStatsSimple(stats *CLIStats) string {
	if stats == nil {
		return "No stats available"
	}

	totalTokens := 0
	totalCost := 0.0

	for modelName, modelStats := range stats.Models {
		totalTokens += modelStats.Tokens.Total
		totalCost += calculateModelCost(modelName, modelStats.Tokens)
	}

	return fmt.Sprintf("Tokens: %d, Tools: %d, Cost: $%.4f",
		totalTokens,
		stats.Tools.TotalCalls,
		totalCost)
}
