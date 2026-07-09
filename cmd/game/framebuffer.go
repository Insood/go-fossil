package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Framebuffer struct {
	Target rl.RenderTexture2D
	Width  int32
	Height int32
}

func NewFramebuffer(width, height int32) *Framebuffer {
	target := rl.RenderTexture2D{}
	target.ID = rl.LoadFramebuffer()
	if target.ID == 0 {
		panic(fmt.Errorf("create framebuffer: rl.LoadFramebuffer returned 0"))
	}

	rl.EnableFramebuffer(target.ID)

	// Create a temporary texture to upload to the GPU
	colorImage := rl.GenImageColor(int(width), int(height), rl.Blank)
	target.Texture = rl.LoadTextureFromImage(colorImage)
	rl.UnloadImage(colorImage)

	target.Depth.ID = rl.LoadTextureDepth(width, height, false)
	target.Depth.Width = width
	target.Depth.Height = height
	target.Depth.Mipmaps = 1                 // not actually used
	target.Depth.Format = rl.UncompressedR32 // not actually used

	rl.FramebufferAttach(target.ID, target.Texture.ID, rl.AttachmentColorChannel0, rl.AttachmentTexture2d, 0)
	rl.FramebufferAttach(target.ID, target.Depth.ID, rl.AttachmentDepth, rl.AttachmentTexture2d, 0)

	complete := rl.FramebufferComplete(target.ID)
	rl.DisableFramebuffer()

	if !complete {
		rl.UnloadRenderTexture(target)
		panic(fmt.Errorf("create framebuffer: framebuffer is incomplete"))
	}

	return &Framebuffer{
		Target: target,
		Width:  width,
		Height: height,
	}
}

func (framebuffer *Framebuffer) Unload() {
	if framebuffer == nil {
		return
	}

	rl.UnloadRenderTexture(framebuffer.Target)
}
