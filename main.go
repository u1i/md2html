package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func main() {
	fontFamily := flag.String("font", "Open Sans", "Google Font family to use")
	title := flag.String("title", "", "HTML document title (empty by default)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: md2html [--font <font-name>] [--title <title>] <input.md>")
		fmt.Println("Example: md2html file.md")
		fmt.Println("Example: md2html --font 'Roboto' file.md")
		fmt.Println("Example: md2html --title 'My Document' file.md")
		os.Exit(1)
	}

	inputFile := args[0]
	if !strings.HasSuffix(inputFile, ".md") {
		fmt.Println("Error: Input file must have .md extension")
		os.Exit(1)
	}

	outputFile := strings.TrimSuffix(inputFile, ".md") + ".html"

	if err := convertMarkdownToHTML(inputFile, outputFile, *fontFamily, *title); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
}

func convertMarkdownToHTML(inputFile, outputFile, fontFamily, title string) error {
	// Read markdown file
	mdContent, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Get the directory of the input file for resolving relative image paths
	inputDir := filepath.Dir(inputFile)

	// Rewrite .md links to .html links
	processedContent := rewriteMarkdownLinks(string(mdContent))

	// Process images and embed them as base64
	processedContent, err = embedImages(processedContent, inputDir)
	if err != nil {
		return fmt.Errorf("failed to process images: %w", err)
	}

	// Convert markdown to HTML
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(processedContent))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlBody := markdown.Render(doc, renderer)

	// Create complete HTML document with Google Font
	fullHTML := createHTMLDocument(string(htmlBody), fontFamily, title)

	// Write to output file
	if err := ioutil.WriteFile(outputFile, []byte(fullHTML), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func rewriteMarkdownLinks(mdContent string) string {
	// Regex to match markdown link syntax: [text](url)
	// But NOT image syntax: ![text](url)
	linkRegex := regexp.MustCompile(`(?m)([^!])\[([^\]]*)\]\(([^)]+)\)`)

	result := linkRegex.ReplaceAllStringFunc(mdContent, func(match string) string {
		// Extract the parts
		parts := linkRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		prefix := parts[1]  // Character before the link (to exclude images)
		linkText := parts[2]
		linkURL := parts[3]

		// Skip if it's an HTTP/HTTPS URL
		if strings.HasPrefix(linkURL, "http://") || 
		   strings.HasPrefix(linkURL, "https://") ||
		   strings.HasPrefix(linkURL, "#") {  // Also skip anchor links
			return match
		}

		// Check if it's a .md file and rewrite to .html
		if strings.HasSuffix(linkURL, ".md") {
			linkURL = strings.TrimSuffix(linkURL, ".md") + ".html"
		}

		return fmt.Sprintf("%s[%s](%s)", prefix, linkText, linkURL)
	})

	// Handle links at the start of a line (no prefix character)
	startLinkRegex := regexp.MustCompile(`(?m)^\[([^\]]*)\]\(([^)]+)\)`)
	result = startLinkRegex.ReplaceAllStringFunc(result, func(match string) string {
		parts := startLinkRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		linkText := parts[1]
		linkURL := parts[2]

		// Skip if it's an HTTP/HTTPS URL or anchor
		if strings.HasPrefix(linkURL, "http://") || 
		   strings.HasPrefix(linkURL, "https://") ||
		   strings.HasPrefix(linkURL, "#") {
			return match
		}

		// Check if it's a .md file and rewrite to .html
		if strings.HasSuffix(linkURL, ".md") {
			linkURL = strings.TrimSuffix(linkURL, ".md") + ".html"
		}

		return fmt.Sprintf("[%s](%s)", linkText, linkURL)
	})

	return result
}

func embedImages(mdContent, baseDir string) (string, error) {
	// Regex to match markdown image syntax: ![alt](path)
	imgRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	result := imgRegex.ReplaceAllStringFunc(mdContent, func(match string) string {
		// Extract the image path
		parts := imgRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		altText := parts[1]
		imgPath := parts[2]

		// Skip if it's already a URL or data URI
		if strings.HasPrefix(imgPath, "http://") || 
		   strings.HasPrefix(imgPath, "https://") || 
		   strings.HasPrefix(imgPath, "data:") {
			return match
		}

		// Resolve relative path
		fullPath := filepath.Join(baseDir, imgPath)

		// Read image file
		imgData, err := ioutil.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("Warning: Could not read image %s: %v\n", imgPath, err)
			return match
		}

		// Detect MIME type based on extension
		mimeType := getMimeType(imgPath)

		// Encode to base64
		base64Data := base64.StdEncoding.EncodeToString(imgData)
		dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

		return fmt.Sprintf("![%s](%s)", altText, dataURI)
	})

	return result, nil
}

func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "image/png" // default
}

func createHTMLDocument(body, fontFamily, title string) string {
	// Convert font family to Google Fonts URL format
	fontURL := strings.ReplaceAll(fontFamily, " ", "+")

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=%s:wght@300;400;600;700&display=swap" rel="stylesheet">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: '%s', sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            background-color: #f9f9f9;
        }
        
        h1, h2, h3, h4, h5, h6 {
            margin-top: 1.5rem;
            margin-bottom: 1rem;
            font-weight: 600;
            line-height: 1.3;
        }
        
        h1 { font-size: 2.5rem; border-bottom: 2px solid #e0e0e0; padding-bottom: 0.5rem; }
        h2 { font-size: 2rem; border-bottom: 1px solid #e0e0e0; padding-bottom: 0.3rem; }
        h3 { font-size: 1.5rem; }
        h4 { font-size: 1.25rem; }
        h5 { font-size: 1.1rem; }
        h6 { font-size: 1rem; }
        
        p {
            margin-bottom: 1rem;
        }
        
        a {
            color: #0066cc;
            text-decoration: none;
        }
        
        a:hover {
            text-decoration: underline;
        }
        
        img {
            max-width: 100%%;
            height: auto;
            display: block;
            margin: 1.5rem auto;
            border-radius: 4px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        
        code {
            background-color: #f4f4f4;
            padding: 0.2rem 0.4rem;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
        }
        
        pre {
            background-color: #f4f4f4;
            padding: 1rem;
            border-radius: 4px;
            overflow-x: auto;
            margin-bottom: 1rem;
        }
        
        pre code {
            background-color: transparent;
            padding: 0;
        }
        
        blockquote {
            border-left: 4px solid #0066cc;
            padding-left: 1rem;
            margin: 1rem 0;
            color: #666;
            font-style: italic;
        }
        
        ul, ol {
            margin-bottom: 1rem;
            padding-left: 2rem;
        }
        
        li {
            margin-bottom: 0.5rem;
        }
        
        table {
            border-collapse: collapse;
            width: 100%%;
            margin-bottom: 1rem;
        }
        
        th, td {
            border: 1px solid #ddd;
            padding: 0.75rem;
            text-align: left;
        }
        
        th {
            background-color: #f4f4f4;
            font-weight: 600;
        }
        
        hr {
            border: none;
            border-top: 2px solid #e0e0e0;
            margin: 2rem 0;
        }
    </style>
</head>
<body>
%s
</body>
</html>`, title, fontURL, fontFamily, body)
}
