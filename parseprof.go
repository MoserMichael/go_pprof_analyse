package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

type CallGraphNode struct {
	Count int           `json:"count,omitempty"`
	Name  string        `json:"name,omitempty"`
	Links []interface{} `json:"called,omitempty"`

	Map     map[string]*CallGraphNode `json:"-"`
	Visited bool                      `json:"-"`
}

type MapTitleToGraphNode map[string]*CallGraphNode

type Results struct {
	MapNameToNode MapTitleToGraphNode
	TopLevelNode  MapTitleToGraphNode
	TopLevelLinks []*CallGraphNode
}

func walkEntry(entry *CallGraphNode) {
	if entry.Visited {
		return
	}
	entry.Visited = true

	for idx, sentry := range entry.Links {

		entryChild, isGraph := sentry.(*CallGraphNode)

		if isGraph {
			if entryChild.Visited {
				entry.Links[idx] = fmt.Sprintf("Backlink: %v", entryChild.Name)
			} else {
				walkEntry(entryChild)
			}
		}
	}
	entry.Visited = false
}

func writeToJson(outFile string, obj interface{}) {
	fileData, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}

	err = os.WriteFile(outFile, fileData, 0644)
	if err != nil {
		log.Fatal("Error writing to file:", err)
	}

}

func formatName(title string, callCount int) string {
	tokens := strings.Split(title[1:], "\t ")
	fmtTitle := strings.Join(tokens, " ")
	return fmt.Sprintf("calls: %v, %v", callCount, fmtTitle)
}

func writeEntry(writer *bufio.Writer, node interface{}) {

	var title string
	valgraph, ok := node.(*CallGraphNode)
	if ok {
		title = formatName(valgraph.Name, valgraph.Count)
	} else {
		valstr, ok := node.(string)
		if ok {
			title = valstr
		}
	}

	if len(valgraph.Links) != 0 {
		writer.WriteString("<details><summary><b>Expand/Collapse</b> " + title + "</summary>\n")
		if valgraph != nil {
			writer.WriteString("<ul>\n")
			for _, entry := range valgraph.Links {
				writer.WriteString("<li>")
				writeEntry(writer, entry)
			}
			writer.WriteString("</ul>\n")
		}
		writer.WriteString("</details>\n")
	} else {
		writer.WriteString(title)
	}
}

func mapToGraphList(mapobj MapTitleToGraphNode) []*CallGraphNode {
	var ret []*CallGraphNode

	for _, val := range mapobj {
		ret = append(ret, val)
	}
	return ret
}


func writeEntries(writer *bufio.Writer, topLevelList []interface{}) {
	//graphVec := mapToGraphList(topLevelNodes)
	//sortByFrequences(graphVec)
	
	for _, node := range topLevelList {
		writeEntry(writer, node)
	}
}

func writeToHtml(outFile string, topLevelList []interface{}) {

	// Create or truncate the file
	file, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close() // Ensure file is closed

	writer := bufio.NewWriter(file)

	writer.WriteString("<html><body>\n")
	writeEntries(writer, topLevelList)
}

func formatRes(res *Results, outFile string) {
	
	graphList := mapToList(res.TopLevelNode)
	sortVec(graphList)

	// check for recursive calls. We can't have recursin in the structure - json marshalling will be a total fail
	for _, entry := range graphList {
		graph := entry.(*CallGraphNode)
		walkEntry(graph)
	}

	//writeToJson(outFile, res.TopLevelNode)
	writeToHtml(outFile, graphList)
}

func mapToList(mapobj MapTitleToGraphNode) []interface{} {
	var ret []interface{}

	for _, val := range mapobj {
		ret = append(ret, val)
	}
	return ret
}

func sortVec(links []interface{}) {
		//fmt.Printf("Node %v len %v len %v\n", entry.Name, len(entry.Links), len(entry.Map))
		sort.Slice(links, func(i, j int) bool {

			v := links[i]
			entry_i, _ := v.(*CallGraphNode)

			v = links[j]
			entry_j, _ := v.(*CallGraphNode)

			return entry_i.Count > entry_j.Count
		})

}

func sortByFrequences(res *Results) {
	for _, entry := range res.MapNameToNode {
		entry.Links = mapToList(entry.Map)
		sortVec(entry.Links)
	}
}

func onScanLine(res *Results, line string, prevLineEntry *CallGraphNode) *CallGraphNode {
	if line == "" || line[0] != '#' {
		return nil
	}
	var entry *CallGraphNode
	var exists bool

	if entry, exists = res.MapNameToNode[line]; !exists {
		entry = &CallGraphNode{Count: 1, Name: line, Map: make(map[string]*CallGraphNode)}
		res.MapNameToNode[line] = entry
	} else {
		entry.Count += 1
	}

	if prevLineEntry != nil {
		_, exists := entry.Map[prevLineEntry.Name]
		if !exists {
			entry.Map[prevLineEntry.Name] = prevLineEntry
		}
	}
	return entry
}

func scanLines(res *Results, fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var prevLineEntry *CallGraphNode

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tmpEntry := onScanLine(res, line, prevLineEntry)
		if tmpEntry == nil && prevLineEntry != nil {
			//fmt.Printf("Add top %v Len %v\n", prevLineEntry.Name, len(prevLineEntry.Map))
			res.TopLevelNode[prevLineEntry.Name] = prevLineEntry
		}
		prevLineEntry = tmpEntry
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error: while scanning %v : %v", fileName, err)
	}
}

func parseCmdLine() (string, string) {
	inFile := flag.String("in", "", "input file - set of outputs of http requests to go process with perf instrumentation")
	outFile := flag.String("out", "out.html", "output file - json html that tracks frequency of each function call, ordered by calling function")

	flag.Parse()

	return *inFile, *outFile
}

func main() {
	inFile, outFile := parseCmdLine()

	res := Results{
		MapNameToNode: make(MapTitleToGraphNode),
		TopLevelNode:  make(MapTitleToGraphNode),
	}

	scanLines(&res, inFile)
	sortByFrequences(&res)
	formatRes(&res, outFile)
}
