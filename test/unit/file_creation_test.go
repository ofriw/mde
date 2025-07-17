package unit

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/ofri/mde/pkg/ast"
)

func TestFileCreation_NonExistingFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "mde-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create editor and try to load a non-existing file
	editor := ast.NewEditor()
	nonExistingFile := filepath.Join(tempDir, "new_file.md")
	
	err = editor.LoadFile(nonExistingFile)
	if err != nil {
		t.Fatalf("Expected no error loading non-existing file, got: %v", err)
	}
	
	// Verify the document was created with empty content
	doc := editor.GetDocument()
	if doc.GetText() != "" {
		t.Errorf("Expected empty content, got: %s", doc.GetText())
	}
	
	// Verify the filename was set correctly
	if doc.GetFilename() != nonExistingFile {
		t.Errorf("Expected filename %s, got: %s", nonExistingFile, doc.GetFilename())
	}
	
	// Verify cursor is at (0,0)
	cursor := editor.GetCursor()
	pos := cursor.GetBufferPos()
	if pos.Line != 0 || pos.Col != 0 {
		t.Errorf("Expected cursor at (0,0), got: (%d,%d)", pos.Line, pos.Col)
	}
}

func TestFileCreation_ExistingFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "mde-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file with content
	testFile := filepath.Join(tempDir, "existing_file.md")
	testContent := "# Test\n\nThis is a test file."
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create editor and load the existing file
	editor := ast.NewEditor()
	err = editor.LoadFile(testFile)
	if err != nil {
		t.Fatalf("Expected no error loading existing file, got: %v", err)
	}
	
	// Verify the content was loaded correctly
	doc := editor.GetDocument()
	if doc.GetText() != testContent {
		t.Errorf("Expected content %s, got: %s", testContent, doc.GetText())
	}
	
	// Verify the filename was set correctly
	if doc.GetFilename() != testFile {
		t.Errorf("Expected filename %s, got: %s", testFile, doc.GetFilename())
	}
}

func TestFileCreation_SaveNewFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "mde-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create editor and load a non-existing file
	editor := ast.NewEditor()
	newFile := filepath.Join(tempDir, "new_file.md")
	
	err = editor.LoadFile(newFile)
	if err != nil {
		t.Fatalf("Expected no error loading non-existing file, got: %v", err)
	}
	
	// Add some content
	editor.InsertText("# New File\n\nThis is a new file created in memory.")
	
	// Save the file
	err = editor.SaveFile("")
	if err != nil {
		t.Fatalf("Expected no error saving file, got: %v", err)
	}
	
	// Verify the file was created on disk
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Errorf("Expected file to be created on disk, but it doesn't exist")
	}
	
	// Verify the content is correct
	content, err := os.ReadFile(newFile)
	if err != nil {
		t.Fatal(err)
	}
	
	expected := "# New File\n\nThis is a new file created in memory."
	if string(content) != expected {
		t.Errorf("Expected content %s, got: %s", expected, string(content))
	}
}