# md2html

A simple Golang CLI tool that converts Markdown files to HTML with embedded images and customizable Google Fonts.

## Features

- ✅ Converts Markdown to clean, styled HTML
- ✅ Embeds images as base64 data URIs (no external dependencies)
- ✅ Uses Google Fonts (default: Open Sans)
- ✅ Customizable font family via command-line flag
- ✅ Customizable HTML title (empty by default)
- ✅ Responsive design with modern styling
- ✅ Supports common Markdown features (headings, lists, code blocks, tables, etc.)

## Installation

```bash
# Clone or navigate to the project directory
cd md2html

# Download dependencies
go mod download

# Build the binary
go build -o md2html

# Optional: Install globally
go install
```

## Usage

### Basic usage (creates file.html with Open Sans font and empty title):
```bash
./md2html file.md
```

### Custom font:
```bash
./md2html --font "Roboto" file.md
```

### Custom title:
```bash
./md2html --title "My Document" file.md
```

### Combine multiple options:
```bash
./md2html --font "Playfair Display" --title "My Beautiful Document" file.md
```

## Examples

Create a sample markdown file:

```bash
cat > example.md << 'EOF'
# Hello World

This is a **markdown** document with some *formatting*.

## Features

- Bullet points
- Multiple items
- Nested lists work too

## Code Example

```go
func main() {
    fmt.Println("Hello, World!")
}
```

## Images

![Sample Image](./image.png)

> This is a blockquote with some wisdom.

EOF
```

Convert it:

```bash
./md2html example.md
# Output: example.html
```

## How It Works

1. **Reads** the input Markdown file
2. **Processes** any local images and converts them to base64 data URIs
3. **Converts** Markdown to HTML using the gomarkdown library
4. **Wraps** the HTML in a complete document with:
   - Google Fonts integration
   - Responsive CSS styling
   - Proper meta tags
5. **Writes** the output to a `.html` file

## Supported Image Formats

- JPEG/JPG
- PNG
- GIF
- BMP
- WebP
- SVG
- ICO

Images referenced via HTTP/HTTPS URLs are left as-is (not embedded).

## Dependencies

- [gomarkdown/markdown](https://github.com/gomarkdown/markdown) - Markdown parser and HTML renderer

## License

MIT
