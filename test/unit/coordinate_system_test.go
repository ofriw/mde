// Package test provides comprehensive tests for the unified coordinate system.
//
// COORDINATE SYSTEM ARCHITECTURE TESTS:
//
// This test suite validates the new Single Source of Truth coordinate system
// implemented in ticket 010-coordinate-system-architecture-refactor.
//
// DESIGN PRINCIPLES BEING TESTED:
// 1. Single Source of Truth: BufferPos is authoritative
// 2. Immutable Configuration: Viewport configuration is immutable
// 3. Unidirectional Flow: BufferPos → Viewport → ScreenPos
// 4. Explicit Error Handling: All operations return clear errors
// 5. Lazy Calculation: Screen positions computed on-demand
//
// TEST CATEGORIES:
// - BufferPos validation and bounds checking
// - Viewport transformations (BufferPos → ScreenPos)
// - CursorManager state management
// - Edge cases: empty documents, viewport boundaries, unicode
// - Error handling: invalid positions, out-of-bounds, not visible
//
// CRITICAL EDGE CASES:
// - Position (0,0) in empty document
// - Cursor at end-of-line positions
// - Viewport scrolling with cursor off-screen
// - Line numbers enabled/disabled transitions
// - Unicode character boundaries
// - Large document performance
package unit

import (
	"testing"

	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBufferPos_Basic validates basic BufferPos functionality
func TestBufferPos_Basic(t *testing.T) {
	t.Run("valid_position", func(t *testing.T) {
		pos := ast.BufferPos{Line: 5, Col: 10}
		assert.True(t, pos.IsValid(), "Valid position should return true")
		assert.Equal(t, "BufferPos{Line:5, Col:10}", pos.String())
	})

	t.Run("invalid_negative_line", func(t *testing.T) {
		pos := ast.BufferPos{Line: -1, Col: 0}
		assert.False(t, pos.IsValid(), "Negative line should be invalid")
	})

	t.Run("invalid_negative_col", func(t *testing.T) {
		pos := ast.BufferPos{Line: 0, Col: -1}
		assert.False(t, pos.IsValid(), "Negative column should be invalid")
	})

	t.Run("zero_position", func(t *testing.T) {
		pos := ast.BufferPos{Line: 0, Col: 0}
		assert.True(t, pos.IsValid(), "Zero position should be valid")
	})
}

// TestViewport_Transformation validates BufferPos → ScreenPos transformations
func TestViewport_Transformation(t *testing.T) {
	t.Run("basic_transformation_no_line_numbers", func(t *testing.T) {
		// Test Case: Basic transformation without line numbers
		// BufferPos(2,5) with viewport at (0,0) should become ScreenPos(2,5)
		viewport := ast.NewViewport(0, 0, 80, 24, 0, 4)
		bufferPos := ast.BufferPos{Line: 2, Col: 5}
		
		screenPos, err := viewport.BufferToScreen(bufferPos)
		require.NoError(t, err, "Transformation should succeed")
		
		assert.Equal(t, 2, screenPos.Row, "Screen row should match buffer line")
		assert.Equal(t, 5, screenPos.Col, "Screen col should match buffer col (no line numbers)")
	})

	t.Run("basic_transformation_with_line_numbers", func(t *testing.T) {
		// Test Case: Basic transformation with line numbers
		// BufferPos(2,5) with viewport at (0,0) and line numbers should become ScreenPos(2,11)
		// Line number prefix adds 6 characters: "   3 │ "
		viewport := ast.NewViewport(0, 0, 80, 24, 6, 4)
		bufferPos := ast.BufferPos{Line: 2, Col: 5}
		
		screenPos, err := viewport.BufferToScreen(bufferPos)
		require.NoError(t, err, "Transformation should succeed")
		
		assert.Equal(t, 2, screenPos.Row, "Screen row should match buffer line")
		assert.Equal(t, 11, screenPos.Col, "Screen col should be buffer col + line number width (5 + 6)")
	})

	t.Run("viewport_scrolling_vertical", func(t *testing.T) {
		// Test Case: Viewport scrolled down by 10 lines
		// BufferPos(15,3) with viewport top at 10 should become ScreenPos(5,3)
		viewport := ast.NewViewport(10, 0, 80, 24, 0, 4)
		bufferPos := ast.BufferPos{Line: 15, Col: 3}
		
		screenPos, err := viewport.BufferToScreen(bufferPos)
		require.NoError(t, err, "Transformation should succeed")
		
		assert.Equal(t, 5, screenPos.Row, "Screen row should be buffer line - viewport top (15 - 10)")
		assert.Equal(t, 3, screenPos.Col, "Screen col should match buffer col")
	})

	t.Run("viewport_scrolling_horizontal", func(t *testing.T) {
		// Test Case: Viewport scrolled right by 20 columns
		// BufferPos(5,25) with viewport left at 20 should become ScreenPos(5,5)
		viewport := ast.NewViewport(0, 20, 80, 24, 0, 4)
		bufferPos := ast.BufferPos{Line: 5, Col: 25}
		
		screenPos, err := viewport.BufferToScreen(bufferPos)
		require.NoError(t, err, "Transformation should succeed")
		
		assert.Equal(t, 5, screenPos.Row, "Screen row should match buffer line")
		assert.Equal(t, 5, screenPos.Col, "Screen col should be buffer col - viewport left (25 - 20)")
	})

	t.Run("viewport_scrolling_with_line_numbers", func(t *testing.T) {
		// Test Case: Combined scrolling with line numbers
		// BufferPos(15,25) with viewport at (10,20) and line numbers should become ScreenPos(5,11)
		// Screen col = (25 - 20) + 6 = 11
		viewport := ast.NewViewport(10, 20, 80, 24, 6, 4)
		bufferPos := ast.BufferPos{Line: 15, Col: 25}
		
		screenPos, err := viewport.BufferToScreen(bufferPos)
		require.NoError(t, err, "Transformation should succeed")
		
		assert.Equal(t, 5, screenPos.Row, "Screen row should be buffer line - viewport top (15 - 10)")
		assert.Equal(t, 11, screenPos.Col, "Screen col should be (buffer col - viewport left) + line number width")
	})
}

// TestViewport_VisibilityChecks validates position visibility logic
func TestViewport_VisibilityChecks(t *testing.T) {
	t.Run("position_not_visible_above_viewport", func(t *testing.T) {
		// Test Case: Position above viewport should not be visible
		viewport := ast.NewViewport(10, 0, 80, 24, 0, 4)
		bufferPos := ast.BufferPos{Line: 5, Col: 10}
		
		_, err := viewport.BufferToScreen(bufferPos)
		assert.Error(t, err, "Position above viewport should not be visible")
		assert.Equal(t, ast.ErrPositionNotVisible, err, "Should return ErrPositionNotVisible")
	})

	t.Run("position_not_visible_below_viewport", func(t *testing.T) {
		// Test Case: Position below viewport should not be visible
		viewport := ast.NewViewport(10, 0, 80, 24, 0, 4)
		bufferPos := ast.BufferPos{Line: 35, Col: 10} // 35 >= 10 + 24
		
		_, err := viewport.BufferToScreen(bufferPos)
		assert.Error(t, err, "Position below viewport should not be visible")
		assert.Equal(t, ast.ErrPositionNotVisible, err, "Should return ErrPositionNotVisible")
	})

	t.Run("position_not_visible_left_of_viewport", func(t *testing.T) {
		// Test Case: Position left of viewport should not be visible
		viewport := ast.NewViewport(0, 10, 80, 24, 0, 4)
		bufferPos := ast.BufferPos{Line: 5, Col: 5} // 5 < 10
		
		_, err := viewport.BufferToScreen(bufferPos)
		assert.Error(t, err, "Position left of viewport should not be visible")
		assert.Equal(t, ast.ErrPositionNotVisible, err, "Should return ErrPositionNotVisible")
	})

	t.Run("position_not_visible_right_of_viewport", func(t *testing.T) {
		// Test Case: Position right of viewport should not be visible
		// With line numbers, visible width is reduced by line number width
		viewport := ast.NewViewport(0, 0, 80, 24, 6, 4)
		bufferPos := ast.BufferPos{Line: 5, Col: 75} // 75 >= 0 + 80 - 6
		
		_, err := viewport.BufferToScreen(bufferPos)
		assert.Error(t, err, "Position right of viewport should not be visible")
		assert.Equal(t, ast.ErrPositionNotVisible, err, "Should return ErrPositionNotVisible")
	})

	t.Run("position_visible_at_viewport_boundary", func(t *testing.T) {
		// Test Case: Position at viewport boundary should be visible
		viewport := ast.NewViewport(10, 20, 80, 24, 6, 4)
		
		testCases := []struct {
			name      string
			bufferPos ast.BufferPos
			expected  ast.ScreenPos
		}{
			{
				name:      "top_left_boundary",
				bufferPos: ast.BufferPos{Line: 10, Col: 20},
				expected:  ast.ScreenPos{Row: 0, Col: 6}, // 20 - 20 + 6
			},
			{
				name:      "bottom_right_boundary",
				bufferPos: ast.BufferPos{Line: 33, Col: 93}, // 33 < 10 + 24, 93 < 20 + 80 - 6
				expected:  ast.ScreenPos{Row: 23, Col: 79}, // 33 - 10, 93 - 20 + 6
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				screenPos, err := viewport.BufferToScreen(tc.bufferPos)
				require.NoError(t, err, "Position at boundary should be visible")
				assert.Equal(t, tc.expected, screenPos, "Screen position should match expected")
			})
		}
	})
}

// TestViewport_Immutability validates immutable viewport operations
func TestViewport_Immutability(t *testing.T) {
	t.Run("with_methods_create_new_viewport", func(t *testing.T) {
		// Test Case: With* methods should create new viewport instances
		original := ast.NewViewport(10, 20, 80, 24, 6, 4)
		
		newTop := original.WithTopLine(15)
		newLeft := original.WithLeftColumn(25)
		newDims := original.WithDimensions(100, 30)
		
		// Original should be unchanged
		assert.Equal(t, 10, original.GetTopLine(), "Original top line should be unchanged")
		assert.Equal(t, 20, original.GetLeftColumn(), "Original left column should be unchanged")
		assert.Equal(t, 80, original.GetWidth(), "Original width should be unchanged")
		assert.Equal(t, 24, original.GetHeight(), "Original height should be unchanged")
		
		// New instances should have updated values
		assert.Equal(t, 15, newTop.GetTopLine(), "New viewport should have updated top line")
		assert.Equal(t, 25, newLeft.GetLeftColumn(), "New viewport should have updated left column")
		assert.Equal(t, 100, newDims.GetWidth(), "New viewport should have updated width")
		assert.Equal(t, 30, newDims.GetHeight(), "New viewport should have updated height")
	})

	t.Run("accessor_methods", func(t *testing.T) {
		// Test Case: All accessor methods should return correct values
		viewport := ast.NewViewport(5, 10, 90, 30, 6, 8)
		
		assert.Equal(t, 5, viewport.GetTopLine(), "GetTopLine should return correct value")
		assert.Equal(t, 10, viewport.GetLeftColumn(), "GetLeftColumn should return correct value")
		assert.Equal(t, 90, viewport.GetWidth(), "GetWidth should return correct value")
		assert.Equal(t, 30, viewport.GetHeight(), "GetHeight should return correct value")
		assert.Equal(t, 6, viewport.GetLineNumberWidth(), "GetLineNumberWidth should return correct value")
		assert.Equal(t, 8, viewport.GetTabWidth(), "GetTabWidth should return correct value")
	})
}

// TestCursorManager_StateManagement validates CursorManager state operations
func TestCursorManager_StateManagement(t *testing.T) {
	// Create a simple mock validator for testing
	validator := &mockValidator{}
	viewport := ast.NewViewport(0, 0, 80, 24, 0, 4)
	
	t.Run("initial_state", func(t *testing.T) {
		// Test Case: CursorManager should start with position (0,0)
		cursor := ast.NewCursorManager(viewport, validator)
		
		pos := cursor.GetBufferPos()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, pos, "Initial position should be (0,0)")
		assert.Equal(t, 0, cursor.GetDesiredColumn(), "Initial desired column should be 0")
		assert.Nil(t, cursor.GetSelection(), "Initial selection should be nil")
		assert.False(t, cursor.HasSelection(), "Should not have selection initially")
	})

	t.Run("set_buffer_position", func(t *testing.T) {
		// Test Case: Setting buffer position should update state
		cursor := ast.NewCursorManager(viewport, validator)
		newPos := ast.BufferPos{Line: 5, Col: 10}
		
		err := cursor.SetBufferPos(newPos)
		require.NoError(t, err, "Setting valid position should succeed")
		
		assert.Equal(t, newPos, cursor.GetBufferPos(), "Position should be updated")
		assert.Equal(t, 10, cursor.GetDesiredColumn(), "Desired column should be updated")
	})

	t.Run("set_position_with_desired_column", func(t *testing.T) {
		// Test Case: Setting position with preserved desired column
		cursor := ast.NewCursorManager(viewport, validator)
		
		// Set initial position and desired column
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 15})
		cursor.SetDesiredColumn(20)
		
		// Move to new position preserving desired column
		newPos := ast.BufferPos{Line: 3, Col: 5}
		err := cursor.SetBufferPosWithDesiredColumn(newPos, true)
		require.NoError(t, err, "Setting position with preserved desired column should succeed")
		
		assert.Equal(t, newPos, cursor.GetBufferPos(), "Position should be updated")
		assert.Equal(t, 20, cursor.GetDesiredColumn(), "Desired column should be preserved")
	})

	t.Run("screen_position_calculation", func(t *testing.T) {
		// Test Case: Screen position should be calculated correctly
		viewport := ast.NewViewport(5, 10, 80, 24, 6, 4)
		cursor := ast.NewCursorManager(viewport, validator)
		
		bufferPos := ast.BufferPos{Line: 10, Col: 20}
		cursor.SetBufferPos(bufferPos)
		
		screenPos, err := cursor.GetScreenPos()
		require.NoError(t, err, "Getting screen position should succeed")
		
		expectedScreenPos := ast.ScreenPos{Row: 5, Col: 16} // (10-5), (20-10+6)
		assert.Equal(t, expectedScreenPos, screenPos, "Screen position should be calculated correctly")
	})

	t.Run("invalid_position_validation", func(t *testing.T) {
		// Test Case: Invalid positions should be rejected
		validator := &mockValidator{shouldFail: true}
		cursor := ast.NewCursorManager(viewport, validator)
		
		invalidPos := ast.BufferPos{Line: -1, Col: 0}
		err := cursor.SetBufferPos(invalidPos)
		
		assert.Error(t, err, "Invalid position should be rejected")
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, cursor.GetBufferPos(), "Position should remain unchanged")
	})
}

// TestCursorManager_SelectionManagement validates selection operations
func TestCursorManager_SelectionManagement(t *testing.T) {
	validator := &mockValidator{}
	viewport := ast.NewViewport(0, 0, 80, 24, 0, 4)
	
	t.Run("start_selection", func(t *testing.T) {
		// Test Case: Starting selection should create selection at current position
		cursor := ast.NewCursorManager(viewport, validator)
		cursor.SetBufferPos(ast.BufferPos{Line: 2, Col: 5})
		
		cursor.StartSelection()
		
		assert.True(t, cursor.HasSelection(), "Should have selection after starting")
		
		selection := cursor.GetSelection()
		require.NotNil(t, selection, "Selection should not be nil")
		assert.Equal(t, ast.BufferPos{Line: 2, Col: 5}, selection.Start, "Selection start should be current position")
		assert.Equal(t, ast.BufferPos{Line: 2, Col: 5}, selection.End, "Selection end should be current position")
	})

	t.Run("extend_selection", func(t *testing.T) {
		// Test Case: Extending selection should update end position
		cursor := ast.NewCursorManager(viewport, validator)
		cursor.SetBufferPos(ast.BufferPos{Line: 2, Col: 5})
		cursor.StartSelection()
		
		// Move cursor and extend selection
		cursor.SetBufferPos(ast.BufferPos{Line: 4, Col: 10})
		cursor.ExtendSelection()
		
		selection := cursor.GetSelection()
		require.NotNil(t, selection, "Selection should not be nil")
		assert.Equal(t, ast.BufferPos{Line: 2, Col: 5}, selection.Start, "Selection start should be unchanged")
		assert.Equal(t, ast.BufferPos{Line: 4, Col: 10}, selection.End, "Selection end should be updated")
	})

	t.Run("clear_selection", func(t *testing.T) {
		// Test Case: Clearing selection should remove selection
		cursor := ast.NewCursorManager(viewport, validator)
		cursor.StartSelection()
		
		cursor.ClearSelection()
		
		assert.False(t, cursor.HasSelection(), "Should not have selection after clearing")
		assert.Nil(t, cursor.GetSelection(), "Selection should be nil after clearing")
	})
}

// TestErrorHandling validates error handling for coordinate operations
func TestErrorHandling(t *testing.T) {
	t.Run("coordinate_errors", func(t *testing.T) {
		// Test Case: Coordinate errors should have proper messages
		bufferPos := ast.BufferPos{Line: 5, Col: 10}
		err := ast.NewBufferCoordinateError(bufferPos, "position out of bounds")
		
		assert.Contains(t, err.Error(), "coordinate error in buffer position", "Error should contain coordinate error message")
		assert.Contains(t, err.Error(), "BufferPos{Line:5, Col:10}", "Error should contain position string")
		assert.Contains(t, err.Error(), "position out of bounds", "Error should contain reason")
	})

	t.Run("screen_coordinate_errors", func(t *testing.T) {
		// Test Case: Screen coordinate errors should have proper messages
		screenPos := ast.ScreenPos{Row: 10, Col: 20}
		err := ast.NewScreenCoordinateError(screenPos, "position outside screen")
		
		assert.Contains(t, err.Error(), "coordinate error in screen position", "Error should contain coordinate error message")
		assert.Contains(t, err.Error(), "ScreenPos{Row:10, Col:20}", "Error should contain position string")
		assert.Contains(t, err.Error(), "position outside screen", "Error should contain reason")
	})
}

// mockValidator is a simple mock implementation of PositionValidator for testing
type mockValidator struct {
	shouldFail bool
}

func (m *mockValidator) ValidateBufferPos(pos ast.BufferPos) error {
	if m.shouldFail {
		return ast.NewBufferCoordinateError(pos, "mock validation failure")
	}
	if !pos.IsValid() {
		return ast.NewBufferCoordinateError(pos, "invalid position")
	}
	return nil
}