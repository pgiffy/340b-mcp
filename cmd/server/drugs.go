package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tealeg/xlsx/v3"
)

type NDCInfo struct {
	NDC11 string `json:"ndc11"`
}

type NDCRecord struct {
	NDC         string `json:"ndc"`
	DrugName    string `json:"drugName"`
	Strength    string `json:"strength"`
	UnitOfMeasure string `json:"unitOfMeasure"`
	Route       string `json:"route"`
	Manufacturer string `json:"manufacturer"`
	PackageSize string `json:"packageSize"`
	Is340B      bool   `json:"is340b"`
}

var (
	ndcCache    = make(map[string]NDCRecord)
	ndcCacheMux sync.RWMutex
)

type NDCGroup struct {
	NDCList []NDCInfo `json:"ndcList"`
}

type NDCResponse struct {
	NDCGroup NDCGroup `json:"ndcGroup"`
}

type RelatedNDCInfo struct {
	NDC11 string `json:"ndc11"`
}

type RelatedNDCResponse struct {
	NDCInfoList struct {
		NDCInfo []RelatedNDCInfo `json:"ndcInfo"`
	} `json:"ndcInfoList"`
}

type RxTermsProperties struct {
	RxCUI     string `json:"rxcui"`
	BrandName string `json:"brandName"`
	FullName  string `json:"fullName"`
	Strength  string `json:"strength"`
	RxTTY     string `json:"rxtty"`
}

type RxTermsResponse struct {
	RxTermsProperties RxTermsProperties `json:"rxtermsProperties"`
}

type ApproximateCandidate struct {
	Name  string `json:"name"`
	RxCUI string `json:"rxcui"`
	Score string `json:"score"`
}

type ApproximateGroup struct {
	Candidate []ApproximateCandidate `json:"candidate"`
}

type ApproximateResponse struct {
	ApproximateGroup ApproximateGroup `json:"approximateGroup"`
}

func CreateRelatedNDCsTool() (mcp.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	tool := mcp.NewTool("get_related_ndcs",
		mcp.WithDescription("Get related NDCs for a given NDC, RxCUI, or drug name"),
		mcp.WithString("ndc",
			mcp.Description("National Drug Code (NDC)"),
		),
		mcp.WithString("rxcui",
			mcp.Description("RxNorm Concept Unique Identifier"),
		),
		mcp.WithString("name",
			mcp.Description("Drug name"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ndc := request.GetString("ndc", "")
		rxcui := request.GetString("rxcui", "")
		name := request.GetString("name", "")

		if ndc == "" && rxcui == "" && name == "" {
			return mcp.NewToolResultError("Missing input: provide ndc, rxcui, or name"), nil
		}

		result, err := getRelatedNDCs(ndc, rxcui, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return tool, handler
}

func CreateRxInfoTool() (mcp.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	tool := mcp.NewTool("get_rx_info",
		mcp.WithDescription("Get detailed information for a given RxCUI"),
		mcp.WithString("rxcui",
			mcp.Required(),
			mcp.Description("RxNorm Concept Unique Identifier"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rxcui, err := request.RequireString("rxcui")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result, err := getRxInfo(rxcui)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return tool, handler
}

func Create340BEligibilityTool() (mcp.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	tool := mcp.NewTool("check_340b_eligibility",
		mcp.WithDescription("Check if a drug is 340B eligible based on NDC, RxCUI, or drug name"),
		mcp.WithString("ndc",
			mcp.Description("National Drug Code (NDC)"),
		),
		mcp.WithString("rxcui",
			mcp.Description("RxNorm Concept Unique Identifier"),
		),
		mcp.WithString("name",
			mcp.Description("Drug name"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ndc := request.GetString("ndc", "")
		rxcui := request.GetString("rxcui", "")
		name := request.GetString("name", "")

		if ndc == "" && rxcui == "" && name == "" {
			return mcp.NewToolResultError("Missing input: provide ndc, rxcui, or name"), nil
		}

		result, err := check340BEligibility(ndc, rxcui, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return tool, handler
}

func CreateApproximateMatchTool() (mcp.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	tool := mcp.NewTool("find_approximate_drug_match",
		mcp.WithDescription("Find approximate drug name matches using RxNorm API"),
		mcp.WithString("term",
			mcp.Required(),
			mcp.Description("Drug name to search for"),
		),
		mcp.WithNumber("max_entries",
			mcp.Description("Maximum number of results to return (default: 1)"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		term, err := request.RequireString("term")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		maxEntries := request.GetFloat("max_entries", 1)
		if maxEntries == 0 {
			maxEntries = 1
		}

		result, err := findApproximateMatch(term, int(maxEntries))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return tool, handler
}

func CreateGenerateRxNormExcelTool() (mcp.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	tool := mcp.NewTool("generate_rxnorm_excel",
		mcp.WithDescription("Process drug names and generate Excel file with RxNorm matches"),
		mcp.WithString("drug_names",
			mcp.Required(),
			mcp.Description("JSON array of drug names to process"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		drugNamesJson, err := request.RequireString("drug_names")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var drugNames []string
		if err := json.Unmarshal([]byte(drugNamesJson), &drugNames); err != nil {
			return mcp.NewToolResultError("Invalid JSON format for drug_names"), nil
		}

		result, err := generateRxNormExcel(drugNames)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return tool, handler
}

func CreateIs340BExcelTool() (mcp.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	tool := mcp.NewTool("is_340b_excel",
		mcp.WithDescription("Process NDC codes and generate Excel file with 340B eligibility status"),
		mcp.WithString("ndc_codes",
			mcp.Required(),
			mcp.Description("JSON array of NDC codes to check"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ndcCodesJson, err := request.RequireString("ndc_codes")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var ndcCodes []string
		if err := json.Unmarshal([]byte(ndcCodesJson), &ndcCodes); err != nil {
			return mcp.NewToolResultError("Invalid JSON format for ndc_codes"), nil
		}

		result, err := generate340BExcel(ndcCodes)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonResult, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return tool, handler
}

func getRelatedNDCs(ndc, rxcui, name string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	if rxcui != "" {
		resp, err := http.Get(fmt.Sprintf("https://rxnav.nlm.nih.gov/REST/rxcui/%s/ndcs.json", rxcui))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		
		var ndcResponse NDCResponse
		if err := json.Unmarshal(body, &ndcResponse); err != nil {
			return nil, err
		}
		
		result["ndcs"] = ndcResponse.NDCGroup.NDCList
	} else if name != "" {
		resp, err := http.Get(fmt.Sprintf("https://rxnav.nlm.nih.gov/REST/rxcui.json?name=%s&search=2", url.QueryEscape(name)))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		
		result["rxcui_search"] = string(body)
	} else if ndc != "" {
		resp, err := http.Get(fmt.Sprintf("https://rxnav.nlm.nih.gov/REST/relatedndc.json?relation=drug&ndc=%s", ndc))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		
		var relatedResponse RelatedNDCResponse
		if err := json.Unmarshal(body, &relatedResponse); err != nil {
			return nil, err
		}
		
		result["related_ndcs"] = relatedResponse.NDCInfoList.NDCInfo
	}
	
	return result, nil
}

func getRxInfo(rxcui string) (map[string]interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("https://rxnav.nlm.nih.gov/REST/RxTerms/rxcui/%s/allinfo.json", rxcui))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var rxResponse RxTermsResponse
	if err := json.Unmarshal(body, &rxResponse); err != nil {
		return nil, err
	}
	
	result := map[string]interface{}{
		"success": true,
		"info":    rxResponse.RxTermsProperties,
	}
	
	return result, nil
}

func check340BEligibility(ndc, rxcui, name string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"is_340b":     false,
		"cache_info":  "Loaded on startup",
	}
	
	var ndcsToCheck []string
	
	if ndc != "" {
		ndcsToCheck = append(ndcsToCheck, ndc)
	}
	
	if rxcui != "" || name != "" {
		relatedNDCs, err := getRelatedNDCs(ndc, rxcui, name)
		if err != nil {
			return nil, err
		}
		
		result["related_ndcs"] = relatedNDCs
		
		if ndcList, ok := relatedNDCs["ndcs"].([]NDCInfo); ok {
			for _, ndcInfo := range ndcList {
				ndcsToCheck = append(ndcsToCheck, ndcInfo.NDC11)
			}
		}
		
		if relatedNDCList, ok := relatedNDCs["related_ndcs"].([]RelatedNDCInfo); ok {
			for _, ndcInfo := range relatedNDCList {
				ndcsToCheck = append(ndcsToCheck, ndcInfo.NDC11)
			}
		}
	}
	
	var eligibleNDCs []NDCRecord
	for _, checkNDC := range ndcsToCheck {
		if record, exists := getNDCFromCache(checkNDC); exists {
			if record.Is340B {
				eligibleNDCs = append(eligibleNDCs, record)
				result["is_340b"] = true
			}
		}
	}
	
	if len(eligibleNDCs) > 0 {
		result["eligible_ndcs"] = eligibleNDCs
	}
	
	return result, nil
}

func findApproximateMatch(term string, maxEntries int) (map[string]interface{}, error) {
	encodedTerm := url.QueryEscape(strings.TrimSpace(term))
	url := fmt.Sprintf("https://rxnav.nlm.nih.gov/REST/approximateTerm.json?term=%s&maxEntries=%d", encodedTerm, maxEntries)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var approxResponse ApproximateResponse
	if err := json.Unmarshal(body, &approxResponse); err != nil {
		return nil, err
	}
	
	result := map[string]interface{}{
		"term":      term,
		"matches":   approxResponse.ApproximateGroup.Candidate,
		"success":   true,
	}
	
	return result, nil
}

func InitNDCCache() error {
	return downloadAndProcessNDCs()
}

func StopNDCCache() {
	// No cleanup needed for local-only version
}

func downloadAndProcessNDCs() error {
	resp, err := http.Get("https://www.340besp.com/ndcs")
	if err != nil {
		return fmt.Errorf("failed to download NDC file: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read NDC file: %v", err)
	}
	
	file, err := xlsx.OpenBinary(body)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %v", err)
	}
	
	if len(file.Sheets) == 0 {
		return fmt.Errorf("no sheets found in Excel file")
	}
	
	sheet := file.Sheets[0]
	newCache := make(map[string]NDCRecord)
	
	// Skip header row, start from row 1
	for i := 1; i < sheet.MaxRow; i++ {
		row, err := sheet.Row(i)
		if err != nil {
			continue
		}
		
		// Get cells manually by column
		ndc := getCellValueByIndex(row, 0)
		drugName := getCellValueByIndex(row, 1)
		strength := getCellValueByIndex(row, 2)
		unitOfMeasure := getCellValueByIndex(row, 3)
		route := getCellValueByIndex(row, 4)
		manufacturer := getCellValueByIndex(row, 5)
		packageSize := getCellValueByIndex(row, 6)
		is340BStr := getCellValueByIndex(row, 7)
		
		record := NDCRecord{
			NDC:           ndc,
			DrugName:      drugName,
			Strength:      strength,
			UnitOfMeasure: unitOfMeasure,
			Route:         route,
			Manufacturer:  manufacturer,
			PackageSize:   packageSize,
			Is340B:        strings.ToLower(is340BStr) == "true" || is340BStr == "1",
		}
		
		if record.NDC != "" {
			newCache[record.NDC] = record
		}
	}
	
	ndcCacheMux.Lock()
	ndcCache = newCache
	ndcCacheMux.Unlock()
	
	fmt.Fprintf(os.Stderr, "NDC cache updated with %d records\n", len(newCache))
	return nil
}

func getCellValue(cell *xlsx.Cell) string {
	if cell == nil {
		return ""
	}
	return strings.TrimSpace(cell.String())
}

func getCellValueByIndex(row *xlsx.Row, index int) string {
	cell := row.GetCell(index)
	return getCellValue(cell)
}

func getNDCFromCache(ndc string) (NDCRecord, bool) {
	ndcCacheMux.RLock()
	defer ndcCacheMux.RUnlock()
	
	record, exists := ndcCache[ndc]
	return record, exists
}

func generateRxNormExcel(drugNames []string) (map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)
	
	for _, drugName := range drugNames {
		drugName = strings.TrimSpace(drugName)
		if drugName == "" {
			continue
		}
		
		// Find approximate matches
		matches, err := findApproximateMatch(drugName, 3) // Get top 3 matches
		if err != nil {
			results = append(results, map[string]interface{}{
				"original_name": drugName,
				"found_name":   "",
				"rxcui":        "",
				"score":        "",
				"error":        err.Error(),
			})
			continue
		}
		
		if candidates, ok := matches["matches"].([]ApproximateCandidate); ok && len(candidates) > 0 {
			// Take the best match
			bestMatch := candidates[0]
			results = append(results, map[string]interface{}{
				"original_name": drugName,
				"found_name":   bestMatch.Name,
				"rxcui":        bestMatch.RxCUI,
				"score":        bestMatch.Score,
				"error":        "",
			})
		} else {
			results = append(results, map[string]interface{}{
				"original_name": drugName,
				"found_name":   "",
				"rxcui":        "",
				"score":        "",
				"error":        "No matches found",
			})
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"results": results,
		"total_processed": len(drugNames),
		"note": "Excel-like data structure for RxNorm matches",
	}, nil
}

func generate340BExcel(ndcCodes []string) (map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)
	
	for _, ndc := range ndcCodes {
		ndc = strings.TrimSpace(ndc)
		if ndc == "" {
			continue
		}
		
		// Check 340B eligibility
		eligibilityResult, err := check340BEligibility(ndc, "", "")
		if err != nil {
			results = append(results, map[string]interface{}{
				"ndc":          ndc,
				"is_340b":      false,
				"drug_name":    "",
				"manufacturer": "",
				"error":        err.Error(),
			})
			continue
		}
		
		is340B := false
		drugName := ""
		manufacturer := ""
		
		if eligibilityResult["is_340b"] == true {
			is340B = true
			
			// Get additional info from cache if available
			if record, exists := getNDCFromCache(ndc); exists {
				drugName = record.DrugName
				manufacturer = record.Manufacturer
			}
		}
		
		results = append(results, map[string]interface{}{
			"ndc":          ndc,
			"is_340b":      is340B,
			"drug_name":    drugName,
			"manufacturer": manufacturer,
			"error":        "",
		})
	}
	
	return map[string]interface{}{
		"success": true,
		"results": results,
		"total_processed": len(ndcCodes),
		"cache_info": "Loaded on startup",
		"note": "Excel-like data structure for 340B eligibility",
	}, nil
}