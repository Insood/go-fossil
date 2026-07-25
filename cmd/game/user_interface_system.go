package main

import (
	"fmt"

	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type UserInterfaceSystem struct {
	filter          *ecs.Filter2[DroneFireControl, Battery]
	artifactManager *ArtifactManager
}

func (system *UserInterfaceSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[DroneFireControl, Battery](game.world)
	system.artifactManager = game.artifactManager
}

func (system *UserInterfaceSystem) Update(game *Game) {
	system.drawTotalScore(game)
	system.drawDroneStatusBars()
	system.drawDroneViewport(game)
	system.drawDroneReticle()
	system.drawArtifactFragments()
}

func (system *UserInterfaceSystem) Unload() {}

func (system *UserInterfaceSystem) drawTotalScore(game *Game) {
	rl.DrawText(fmt.Sprintf("Score: %d", game.TotalScore), totalScoreTextX, totalScoreTextY, 24, rl.White)
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

func (system *UserInterfaceSystem) drawDroneStatusBars() {
	query := system.filter.Query()
	defer query.Close()

	batteryPercent := float32(0)
	for query.Next() {
		_, battery := query.Get()
		batteryPercent = battery.charge / droneBatteryCharge
		break
	}

	viewport := droneViewportRectangle()
	drawDroneStatusBar("CARGO", droneCargoPercent(system.artifactManager), rl.Gold, viewport, 1)
	drawDroneStatusBar("BATTERY", batteryPercent, rl.Lime, viewport, 0)
}

func drawDroneStatusBar(label string, fillPercent float32, fillColor rl.Color, viewport rl.Rectangle, row int32) {
	barX, barY, barWidth, labelX, labelY := droneStatusBarLayout(viewport, row)
	innerWidth := barWidth - droneBatteryBarPadding*2
	fillWidth := int32(float32(innerWidth) * clampFloat32(fillPercent, 0, 1))

	rl.DrawText(label, labelX, labelY, droneBatteryLabelFontSize, rl.White)
	rl.DrawRectangle(barX, barY, barWidth, droneBatteryBarHeight, rl.Fade(rl.Black, 0.8))
	rl.DrawRectangle(
		barX+droneBatteryBarPadding,
		barY+droneBatteryBarPadding,
		fillWidth,
		droneBatteryBarHeight-droneBatteryBarPadding*2,
		fillColor,
	)
}

func droneStatusBarLayout(viewport rl.Rectangle, row int32) (barX, barY, barWidth, labelX, labelY int32) {
	labelWidth := rl.MeasureText("BATTERY", droneBatteryLabelFontSize)
	labelX = int32(viewport.X)
	barY = int32(viewport.Y) -
		droneBatteryBarGap -
		droneBatteryBarHeight -
		row*(droneBatteryBarHeight+droneBatteryBarGap)
	labelY = barY + (droneBatteryBarHeight-droneBatteryLabelFontSize)/2
	barX = labelX + labelWidth + droneBatteryLabelGap
	barWidth = int32(viewport.Width) - labelWidth - droneBatteryLabelGap + droneBatteryBarExtraWidth
	return barX, barY, barWidth, labelX, labelY
}

func droneCargoPercent(manager *ArtifactManager) float32 {
	return clampFloat32(
		float32(manager.CarriedFragmentWeight())/float32(droneMaximumCarryWeight),
		0,
		1,
	)
}

func (system *UserInterfaceSystem) drawDroneReticle() {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		control, _ := query.Get()
		center := droneViewportCursorPixel(control.cursor, droneViewportRectangle())

		rl.DrawLineEx(
			rl.Vector2{X: center.X - droneViewReticleHalfSize, Y: center.Y},
			rl.Vector2{X: center.X + droneViewReticleHalfSize, Y: center.Y},
			droneViewReticleThickness,
			rl.Red,
		)
		rl.DrawLineEx(
			rl.Vector2{X: center.X, Y: center.Y - droneViewReticleHalfSize},
			rl.Vector2{X: center.X, Y: center.Y + droneViewReticleHalfSize},
			droneViewReticleThickness,
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
		scoreText := fmt.Sprintf("%d", fragment.Score)
		textY := int32(rowY + 22)
		rl.DrawText(scoreText, int32(textX), textY, 20, rl.White)

		gradeText, gradeColor := artifactFragmentGradeDisplay(fragment.Grade)
		gradeX := int32(textX) + rl.MeasureText(scoreText, 20) + artifactFragmentGradeGap
		rl.DrawText(gradeText, gradeX, textY, 20, gradeColor)
	}
}

func artifactFragmentGradeDisplay(grade float64) (string, rl.Color) {
	switch {
	case grade >= 1:
		return "[SUPER]", rl.Gold
	case grade >= 0.9:
		return "[A]", rl.Green
	case grade >= 0.8:
		return "[B]", rl.Yellow
	case grade >= 0.7:
		return "[C]", rl.Orange
	default:
		return "[F]", rl.Red
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
		if !fragment.Collected || fragment.DroppedOff {
			continue
		}
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
