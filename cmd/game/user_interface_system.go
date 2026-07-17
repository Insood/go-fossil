package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
	"slices"
)

type UserInterfaceSystem struct {
	filter          *ecs.Filter1[DroneFireControl]
	artifactManager *ArtifactManager
}

func (system *UserInterfaceSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter1[DroneFireControl](game.world)
	system.artifactManager = game.artifactManager
}

func (system *UserInterfaceSystem) Update(game *Game) {
	system.drawTotalScore(game)
	system.drawDroneViewport(game)
	system.drawDroneReticle()
	system.drawArtifactFragments()
}

func (system *UserInterfaceSystem) drawTotalScore(game *Game) {
	rl.DrawText(fmt.Sprintf("Total Score: %d", game.TotalScore), totalScoreTextX, totalScoreTextY, 24, rl.White)
}

func (system *UserInterfaceSystem) drawDroneViewport(game *Game) {
	viewport := droneViewportRectangle()

	rl.DrawRectangle(
		int32(viewport.X-2),
		int32(viewport.Y-2),
		droneViewPixels+4,
		droneViewPixels+4,
		rl.Fade(rl.Black, 0.8),
	)

	rl.DrawTexturePro(
		game.droneFramebuffer.Target.Texture,
		rl.NewRectangle(0, 0, float32(game.droneFramebuffer.Width), -float32(game.droneFramebuffer.Height)),
		viewport,
		rl.Vector2{},
		0,
		rl.White,
	)
}

func (system *UserInterfaceSystem) drawDroneReticle() {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		control := query.Get()
		center := control.cursor
		halfSize := int32(5)

		rl.DrawLine(
			int32(center.X)-halfSize,
			int32(center.Y),
			int32(center.X)+halfSize,
			int32(center.Y),
			rl.Red,
		)
		rl.DrawLine(
			int32(center.X),
			int32(center.Y)-halfSize,
			int32(center.X),
			int32(center.Y)+halfSize,
			rl.Red,
		)
	}
}

func (system *UserInterfaceSystem) drawArtifactFragments() {
	fragments := recentArtifactFragments(system.artifactManager, artifactFragmentDisplayCount)
	if len(fragments) == 0 {
		return
	}

	fragmentStartY := screenHeight - artifactFragmentStartY - artifactFragmentThumbSize

	for i, fragment := range fragments {
		rowY := float32(fragmentStartY - i*artifactFragmentRowStep)
		drawArtifactFragmentThumbnail(fragment, float32(artifactFragmentStartX), rowY, artifactFragmentThumbSize)

		textX := float32(artifactFragmentStartX + artifactFragmentThumbSize + artifactFragmentTextGap)
		rl.DrawText(fmt.Sprintf("%d", fragment.Weight), int32(textX), int32(rowY+6), 20, rl.White)
		rl.DrawText(fmt.Sprintf("%d", fragment.Score), int32(textX), int32(rowY+32), 20, rl.White)
	}
}

func drawArtifactFragmentThumbnail(fragment *ArtifactFragment, x, y float32, size int32) {
	if fragment == nil || fragment.Texture.ID == 0 || fragment.Image == nil {
		return
	}

	bounds := fragment.Image.Bounds()
	if bounds.Empty() {
		return
	}

	sourceWidth := float32(bounds.Dx())
	sourceHeight := float32(bounds.Dy())
	scale := minFloat32(float32(size)/sourceWidth, float32(size)/sourceHeight)
	drawWidth := float32(bounds.Dx()) * scale
	drawHeight := float32(bounds.Dy()) * scale
	destX := x + (float32(size)-drawWidth)/2
	destY := y + (float32(size)-drawHeight)/2

	rl.DrawTexturePro(
		fragment.Texture,
		rl.NewRectangle(0, 0, sourceWidth, sourceHeight),
		rl.NewRectangle(destX, destY, drawWidth, drawHeight),
		rl.Vector2{},
		0,
		rl.White,
	)
}

func sortedArtifactFragments(manager *ArtifactManager) []*ArtifactFragment {
	if manager == nil || len(manager.fragments) == 0 {
		return nil
	}

	fragments := make([]*ArtifactFragment, 0, len(manager.fragments))
	for _, fragment := range manager.fragments {
		fragments = append(fragments, fragment)
	}

	slices.SortFunc(fragments, func(left, right *ArtifactFragment) int {
		switch {
		case left.ID < right.ID:
			return -1
		case left.ID > right.ID:
			return 1
		default:
			return 0
		}
	})

	return fragments
}

func recentArtifactFragments(manager *ArtifactManager, limit int) []*ArtifactFragment {
	fragments := sortedArtifactFragments(manager)
	if len(fragments) == 0 || limit <= 0 {
		return nil
	}
	if len(fragments) > limit {
		fragments = fragments[len(fragments)-limit:]
	}

	recent := make([]*ArtifactFragment, 0, len(fragments))
	for i := len(fragments) - 1; i >= 0; i-- {
		recent = append(recent, fragments[i])
	}

	return recent
}
