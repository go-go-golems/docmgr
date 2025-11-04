package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/docmgr/docmgr/pkg/models"
	"github.com/docmgr/docmgr/pkg/utils"
	"gopkg.in/yaml.v3"
)

type Server struct {
	rootDir string
}

type InitRequest struct {
	Ticket string   `json:"ticket"`
	Title  string   `json:"title"`
	Topics []string `json:"topics"`
}

type AddDocRequest struct {
	Ticket  string `json:"ticket"`
	DocType string `json:"docType"`
	Title   string `json:"title"`
}

type ImportFileRequest struct {
	Ticket   string `json:"ticket"`
	FileName string `json:"fileName"`
	Content  string `json:"content"`
	Name     string `json:"name"`
}

func NewServer(rootDir string) *Server {
	return &Server{rootDir: rootDir}
}

func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create slug from title
	slug := utils.Slugify(req.Title)
	dirName := fmt.Sprintf("%s-%s", req.Ticket, slug)
	ticketPath := filepath.Join(s.rootDir, "active", dirName)

	// Create directory structure
	dirs := []string{
		ticketPath,
		filepath.Join(ticketPath, "design"),
		filepath.Join(ticketPath, "reference"),
		filepath.Join(ticketPath, "playbooks"),
		filepath.Join(ticketPath, "scripts"),
		filepath.Join(ticketPath, "sources"),
		filepath.Join(ticketPath, ".meta"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Create index.md with frontmatter
	doc := models.Document{
		Title:           req.Title,
		Ticket:          req.Ticket,
		Status:          "active",
		Topics:          req.Topics,
		DocType:         "index",
		Intent:          "long-term",
		Owners:          []string{},
		RelatedFiles:    models.RelatedFiles{},
		ExternalSources: []string{},
		Summary:         "",
		LastUpdated:     time.Now(),
	}

	indexPath := filepath.Join(ticketPath, "index.md")
	content := fmt.Sprintf("# %s\n\nDocument workspace for %s.\n", req.Title, req.Ticket)
	if err := writeDocumentWithFrontmatter(indexPath, &doc, content); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write index.md: %v", err), http.StatusInternalServerError)
		return
	}

	// Create README.md
	readmePath := filepath.Join(ticketPath, "README.md")
	readmeContent := fmt.Sprintf(`# %s

This is the document workspace for ticket %s.

## Structure

- **design/**: Design documents and architecture notes
- **reference/**: Reference documentation and API contracts
- **playbooks/**: Operational playbooks and procedures
- **scripts/**: Utility scripts and automation
- **sources/**: External sources and imported documents
`, req.Title, req.Ticket)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write README.md: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ticket": req.Ticket,
		"path":   ticketPath,
		"title":  req.Title,
		"status": "created",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activePath := filepath.Join(s.rootDir, "active")
	if _, err := os.Stat(activePath); os.IsNotExist(err) {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	entries, err := os.ReadDir(activePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read active directory: %v", err), http.StatusInternalServerError)
		return
	}

	var documents []map[string]interface{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		indexPath := filepath.Join(activePath, entry.Name(), "index.md")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			continue
		}

		doc, err := readDocumentFrontmatter(indexPath)
		if err != nil {
			continue
		}

		documents = append(documents, map[string]interface{}{
			"ticket":      doc.Ticket,
			"title":       doc.Title,
			"status":      doc.Status,
			"topics":      doc.Topics,
			"path":        filepath.Join(activePath, entry.Name()),
			"lastUpdated": doc.LastUpdated.Format("2006-01-02"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find the ticket directory
	ticketDir, err := findTicketDirectory(s.rootDir, req.Ticket)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to find ticket: %v", err), http.StatusNotFound)
		return
	}

	// Determine subdirectory based on doc type
	var subdir string
	switch req.DocType {
	case "design-doc":
		subdir = "design"
	case "reference":
		subdir = "reference"
	case "playbook":
		subdir = "playbooks"
	default:
		http.Error(w, "Unknown document type", http.StatusBadRequest)
		return
	}

	// Create filename from title
	slug := utils.Slugify(req.Title)
	filename := fmt.Sprintf("%s.md", slug)
	docPath := filepath.Join(ticketDir, subdir, filename)

	// Check if file already exists
	if _, err := os.Stat(docPath); err == nil {
		http.Error(w, "Document already exists", http.StatusConflict)
		return
	}

	// Read ticket metadata
	indexPath := filepath.Join(ticketDir, "index.md")
	ticketDoc, err := readDocumentFrontmatter(indexPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read ticket metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Create document with frontmatter
	doc := models.Document{
		Title:           req.Title,
		Ticket:          req.Ticket,
		Status:          ticketDoc.Status,
		Topics:          ticketDoc.Topics,
		DocType:         req.DocType,
		Intent:          "long-term",
		Owners:          ticketDoc.Owners,
		RelatedFiles:    models.RelatedFiles{},
		ExternalSources: []string{},
		Summary:         "",
		LastUpdated:     time.Now(),
	}

	content := fmt.Sprintf("# %s\n\n<!-- Add your content here -->\n", req.Title)
	if err := writeDocumentWithFrontmatter(docPath, &doc, content); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write document: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ticket":  req.Ticket,
		"docType": req.DocType,
		"title":   req.Title,
		"path":    docPath,
		"status":  "created",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleImportFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ImportFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find the ticket directory
	ticketDir, err := findTicketDirectory(s.rootDir, req.Ticket)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to find ticket: %v", err), http.StatusNotFound)
		return
	}

	// Create sources directory if it doesn't exist
	sourcesDir := filepath.Join(ticketDir, "sources", "local")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create sources directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Determine destination filename
	destName := req.FileName
	if req.Name != "" {
		ext := filepath.Ext(req.FileName)
		destName = req.Name + ext
	}
	destPath := filepath.Join(sourcesDir, destName)

	// Write the file
	if err := os.WriteFile(destPath, []byte(req.Content), 0644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
		return
	}

	// Create metadata file
	source := models.ExternalSource{
		Type:        "local",
		Path:        req.FileName,
		LastFetched: time.Now(),
	}

	metaPath := filepath.Join(ticketDir, ".meta", "sources.yaml")
	if err := appendSourceMetadata(metaPath, &source); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write metadata: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ticket":      req.Ticket,
		"sourceFile":  req.FileName,
		"destination": destPath,
		"type":        "local",
		"status":      "imported",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeDocumentWithFrontmatter(path string, doc *models.Document, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write frontmatter
	if _, err := f.WriteString("---\n"); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(doc); err != nil {
		return err
	}
	encoder.Close()

	if _, err := f.WriteString("---\n\n"); err != nil {
		return err
	}

	// Write content
	if _, err := f.WriteString(content); err != nil {
		return err
	}

	return nil
}

func readDocumentFrontmatter(path string) (*models.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var doc models.Document
	_, err = frontmatter.Parse(f, &doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func findTicketDirectory(root, ticket string) (string, error) {
	activePath := filepath.Join(root, "active")
	entries, err := os.ReadDir(activePath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		indexPath := filepath.Join(activePath, entry.Name(), "index.md")
		doc, err := readDocumentFrontmatter(indexPath)
		if err != nil {
			continue
		}

		if doc.Ticket == ticket {
			return filepath.Join(activePath, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("ticket not found: %s", ticket)
}

func appendSourceMetadata(path string, source *models.ExternalSource) error {
	var sources []models.ExternalSource

	// Read existing sources if file exists
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(data, &sources); err != nil {
			return err
		}
	}

	sources = append(sources, *source)

	data, err := yaml.Marshal(sources)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	rootDir := os.Getenv("DOCMGR_ROOT")
	if rootDir == "" {
		rootDir = "docs"
	}

	server := NewServer(rootDir)

	http.HandleFunc("/api/init", enableCORS(server.handleInit))
	http.HandleFunc("/api/list", enableCORS(server.handleList))
	http.HandleFunc("/api/add", enableCORS(server.handleAdd))
	http.HandleFunc("/api/import", enableCORS(server.handleImportFile))
	http.HandleFunc("/api/documents", enableCORS(server.handleGetDocuments))
	http.HandleFunc("/api/search", enableCORS(server.handleSearch))
	http.HandleFunc("/api/update", handleUpdateDocument)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting docmgr server on port %s with root directory: %s\n", port, rootDir)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) handleGetDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		http.Error(w, "ticket parameter is required", http.StatusBadRequest)
		return
	}

	// Find the ticket directory
	ticketDir, err := findTicketDirectory(s.rootDir, ticket)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to find ticket: %v", err), http.StatusNotFound)
		return
	}

	var documents []map[string]interface{}

	// Scan design, reference, playbooks, and sources directories
	subdirs := map[string]string{
		"design":    "design",
		"reference": "reference",
		"playbooks": "playbook",
		"sources":   "source",
	}

	for subdir, dt := range subdirs {
		subdirPath := filepath.Join(ticketDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(subdirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(info.Name(), ".md") {
				return nil
			}

			relPath, _ := filepath.Rel(ticketDir, path)
			
			// Try to read frontmatter, but don't fail if it doesn't exist
			var title string
			var docTopics []string
			var docStatus, docIntent, docTypeStr string
			var docOwners []string
			
			doc, err := readDocumentFrontmatter(path)
			if err == nil && doc.Title != "" {
				title = doc.Title
				docTopics = doc.Topics
				docStatus = doc.Status
				docIntent = doc.Intent
				docTypeStr = doc.DocType
				docOwners = doc.Owners
			} else {
				// Use filename without extension as title
				title = strings.TrimSuffix(info.Name(), ".md")
				title = strings.ReplaceAll(title, "-", " ")
				// Capitalize first letter of each word
				words := strings.Fields(title)
				for i, word := range words {
					if len(word) > 0 {
						words[i] = strings.ToUpper(string(word[0])) + word[1:]
					}
				}
				title = strings.Join(words, " ")
				docTopics = []string{}
				docStatus = "draft"
				docIntent = "long-term"
				docTypeStr = dt
				docOwners = []string{}
			}

			var relatedFiles models.RelatedFiles
			var externalSources []string
			var summary string
			if err == nil {
				relatedFiles = doc.RelatedFiles
				externalSources = doc.ExternalSources
				summary = doc.Summary
			}

			documents = append(documents, map[string]interface{}{
				"name":            title,
				"type":            dt,
				"path":            relPath,
				"topics":          docTopics,
				"status":          docStatus,
				"intent":          docIntent,
				"docType":         docTypeStr,
				"owners":          docOwners,
				"summary":         summary,
				"relatedFiles":    relatedFiles,
				"externalSources": externalSources,
			})

			return nil
		})

		if err != nil {
			log.Printf("Error walking directory %s: %v", subdirPath, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	topic := r.URL.Query().Get("topic")
	docType := r.URL.Query().Get("type")

	activePath := filepath.Join(s.rootDir, "active")
	if _, err := os.Stat(activePath); os.IsNotExist(err) {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	entries, err := os.ReadDir(activePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read active directory: %v", err), http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ticketDir := filepath.Join(activePath, entry.Name())
		indexPath := filepath.Join(ticketDir, "index.md")
		
		// Read workspace metadata
		workspaceDoc, err := readDocumentFrontmatter(indexPath)
		if err != nil {
			continue
		}

		// Filter by topic if specified
		if topic != "" {
			hasTopicMatch := false
			for _, t := range workspaceDoc.Topics {
				if strings.EqualFold(t, topic) {
					hasTopicMatch = true
					break
				}
			}
			if !hasTopicMatch {
				continue
			}
		}

		// Scan documents in this workspace
		subdirs := map[string]string{
			"design":    "design",
			"reference": "reference",
			"playbooks": "playbook",
			"sources":   "source",
		}

		for subdir, dt := range subdirs {
			// Filter by document type if specified
			if docType != "" && dt != docType {
				continue
			}

			subdirPath := filepath.Join(ticketDir, subdir)
			if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
				continue
			}

			filepath.Walk(subdirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
					return nil
				}

				relPath, _ := filepath.Rel(ticketDir, path)
				
				var title string
				var docTopics []string
				var status, intent, docType string
				var owners []string
				
				doc, err := readDocumentFrontmatter(path)
				if err == nil && doc.Title != "" {
					title = doc.Title
					docTopics = doc.Topics
					status = doc.Status
					intent = doc.Intent
					docType = doc.DocType
					owners = doc.Owners
				} else {
					title = strings.TrimSuffix(info.Name(), ".md")
					title = strings.ReplaceAll(title, "-", " ")
					words := strings.Fields(title)
					for i, word := range words {
						if len(word) > 0 {
							words[i] = strings.ToUpper(string(word[0])) + word[1:]
						}
					}
					title = strings.Join(words, " ")
					docTopics = workspaceDoc.Topics
					status = "draft"
					intent = "long-term"
					docType = dt
					owners = []string{}
				}

				// Apply search query filter
				if query != "" {
					queryLower := strings.ToLower(query)
					titleLower := strings.ToLower(title)
					pathLower := strings.ToLower(relPath)
					
					if !strings.Contains(titleLower, queryLower) && !strings.Contains(pathLower, queryLower) {
						return nil
					}
				}

					var relatedFiles models.RelatedFiles
					var externalSources []string
					var summary string
					if err == nil {
						relatedFiles = doc.RelatedFiles
						externalSources = doc.ExternalSources
						summary = doc.Summary
					}

					results = append(results, map[string]interface{}{
						"name":            title,
						"type":            dt,
						"path":            relPath,
						"workspace":       workspaceDoc.Ticket,
						"workspaceTitle":  workspaceDoc.Title,
						"topics":          docTopics,
						"status":          status,
						"intent":          intent,
						"docType":         docType,
						"owners":          owners,
						"summary":         summary,
						"relatedFiles":    relatedFiles,
						"externalSources": externalSources,
					})

				return nil
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}


func handleUpdateDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Ticket   string   `json:"ticket"`
		Path     string   `json:"path"`
		Topics   []string `json:"topics"`
		Status   string   `json:"status"`
		Intent   string   `json:"intent"`
		Owners   []string `json:"owners"`
		Summary  string   `json:"summary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read the current document
	docRoot := os.Getenv("DOCMGR_ROOT")
	if docRoot == "" {
		docRoot = "docs"
	}

	fullPath := filepath.Join(docRoot, "active", req.Ticket+"-"+strings.ToLower(strings.ReplaceAll(req.Path, " ", "-")))
	if !strings.HasPrefix(req.Path, req.Ticket) {
		// Path is relative to ticket directory
		fullPath = filepath.Join(docRoot, "active", req.Ticket+"-*", req.Path)
		matches, err := filepath.Glob(fullPath)
		if err != nil || len(matches) == 0 {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		fullPath = matches[0]
	}

	// Read existing content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "Failed to read document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse frontmatter and content
	var doc models.Document
	rest, err := frontmatter.Parse(strings.NewReader(string(content)), &doc)
	if err != nil {
		// No frontmatter exists, create new
		doc = models.Document{}
	}

	// Update metadata
	doc.Topics = req.Topics
	doc.Status = req.Status
	doc.Intent = req.Intent
	doc.Owners = req.Owners
	doc.Summary = req.Summary
	doc.LastUpdated = time.Now()

	// Serialize frontmatter
	var buf bytes.Buffer
	buf.WriteString("---\n")
	
	if doc.Title != "" {
		buf.WriteString(fmt.Sprintf("title: %s\n", doc.Title))
	}
	if doc.Ticket != "" {
		buf.WriteString(fmt.Sprintf("ticket: %s\n", doc.Ticket))
	}
	buf.WriteString(fmt.Sprintf("status: %s\n", doc.Status))
	if len(doc.Topics) > 0 {
		buf.WriteString("topics:\n")
		for _, topic := range doc.Topics {
			buf.WriteString(fmt.Sprintf("  - %s\n", topic))
		}
	}
	if doc.DocType != "" {
		buf.WriteString(fmt.Sprintf("docType: %s\n", doc.DocType))
	}
	buf.WriteString(fmt.Sprintf("intent: %s\n", doc.Intent))
	if len(doc.Owners) > 0 {
		buf.WriteString("owners:\n")
		for _, owner := range doc.Owners {
			buf.WriteString(fmt.Sprintf("  - %s\n", owner))
		}
	}
	if len(doc.RelatedFiles) > 0 {
		buf.WriteString("relatedFiles:\n")
		for _, file := range doc.RelatedFiles {
			buf.WriteString(fmt.Sprintf("  - %s\n", file))
		}
	}
	if len(doc.ExternalSources) > 0 {
		buf.WriteString("externalSources:\n")
		for _, source := range doc.ExternalSources {
			buf.WriteString(fmt.Sprintf("  - %s\n", source))
		}
	}
	if doc.Summary != "" {
		buf.WriteString(fmt.Sprintf("summary: %s\n", doc.Summary))
	}
	buf.WriteString(fmt.Sprintf("lastUpdated: %s\n", doc.LastUpdated.Format(time.RFC3339)))
	buf.WriteString("---\n\n")

	// Append original content (without old frontmatter)
	buf.Write(rest)

	// Write back to file
	if err := os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
		http.Error(w, "Failed to write document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "updated",
		"path":    fullPath,
		"message": "Document metadata updated successfully",
	})
}
