package email

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// StreamingAttachmentProcessor provides memory-efficient attachment processing
type StreamingAttachmentProcessor struct {
	maxSize      int
	maxTotalSize int
}

// NewStreamingAttachmentProcessor creates a new processor with size limits
func NewStreamingAttachmentProcessor(maxSize, maxTotalSize int) *StreamingAttachmentProcessor {
	return &StreamingAttachmentProcessor{
		maxSize:      maxSize,
		maxTotalSize: maxTotalSize,
	}
}

// ValidateBase64Stream validates base64 content without loading it entirely into memory
func (p *StreamingAttachmentProcessor) ValidateBase64Stream(base64Content string) error {
	// Create a reader for the base64 content
	reader := strings.NewReader(base64Content)
	decoder := base64.NewDecoder(base64.StdEncoding, reader)

	// Read in chunks to avoid loading everything into memory
	buffer := make([]byte, 8192) // 8KB chunks
	totalSize := 0

	for {
		n, err := decoder.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("invalid base64 content: %w", err)
		}

		totalSize += n

		// Check size limit during streaming
		if totalSize > p.maxSize {
			return fmt.Errorf("attachment exceeds %dMB limit", p.maxSize/(1024*1024))
		}
	}

	return nil
}

// ProcessAttachmentsStreaming processes attachments with minimal memory footprint
func (p *StreamingAttachmentProcessor) ProcessAttachmentsStreaming(attachments []EmailAttachment) error {
	totalEstimatedSize := 0

	for i, att := range attachments {
		// Quick size estimation first (no memory allocation)
		estimatedSize := p.estimateDecodedSize(att.Content)
		totalEstimatedSize += estimatedSize

		if estimatedSize > p.maxSize {
			return fmt.Errorf("attachment %d (%s) exceeds %dMB limit",
				i+1, att.Filename, p.maxSize/(1024*1024))
		}

		if totalEstimatedSize > p.maxTotalSize {
			return fmt.Errorf("total attachment size exceeds %dMB limit",
				p.maxTotalSize/(1024*1024))
		}

		// Only do streaming validation for larger files (>1MB estimated)
		if estimatedSize > 1024*1024 {
			if err := p.ValidateBase64Stream(att.Content); err != nil {
				return fmt.Errorf("attachment %d (%s): %w", i+1, att.Filename, err)
			}
		} else {
			// For smaller files, use simple validation
			if _, err := base64.StdEncoding.DecodeString(att.Content[:min(100, len(att.Content))]); err != nil {
				return fmt.Errorf("attachment %d (%s): invalid base64 content", i+1, att.Filename)
			}
		}
	}

	return nil
}

// estimateDecodedSize calculates the approximate decoded size without decoding
func (p *StreamingAttachmentProcessor) estimateDecodedSize(base64Content string) int {
	base64Len := len(base64Content)

	// Count padding characters
	padding := 0
	if base64Len > 0 && base64Content[base64Len-1] == '=' {
		padding++
		if base64Len > 1 && base64Content[base64Len-2] == '=' {
			padding++
		}
	}

	// Calculate actual size: (base64_length - padding) * 3 / 4
	return (base64Len - padding) * 3 / 4
}

// min returns the minimum of two integers (Go 1.18+ has this built-in)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
