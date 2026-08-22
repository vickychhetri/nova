package render

import (
	"sync"
)

// CommandBuffer stores a recorded list of drawing commands for a frame or layer.
type CommandBuffer struct {
	mu       sync.Mutex
	commands []Command
}

// NewCommandBuffer creates an empty command buffer.
func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands: make([]Command, 0, 128),
	}
}

// Push adds a drawing command to the buffer.
func (cb *CommandBuffer) Push(cmd Command) {
	cb.commands = append(cb.commands, cmd)
}

// Commands returns a slice of the recorded commands.
func (cb *CommandBuffer) Commands() []Command {
	return cb.commands
}

// Len returns the number of recorded commands.
func (cb *CommandBuffer) Len() int {
	return len(cb.commands)
}

// Clear resets the buffer for reuse without reallocating underlying memory.
func (cb *CommandBuffer) Clear() {
	cb.commands = cb.commands[:0]
}
