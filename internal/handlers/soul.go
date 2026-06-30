package handlers

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hurtener/dockyard/runtime/tool"
	"github.com/hurtener/pptx-go/pptx"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

func (h *handlers) bootstrapSoul(_ context.Context, in contracts.BootstrapSoulInput) (tool.Result[contracts.BootstrapSoulOutput], error) {
	var palette *soul.Palette
	if in.Palette != nil {
		palette = &soul.Palette{
			Surfaces:   in.Palette.Surfaces,
			Text:       in.Palette.Text,
			Extensions: in.Palette.Extensions,
		}
	}
	var darkPalette *soul.DarkPalette
	if in.DarkPalette != nil {
		darkPalette = &soul.DarkPalette{
			DarkSurfaces: in.DarkPalette.DarkSurfaces,
			DarkText:     in.DarkPalette.DarkText,
		}
	}
	bootstrapped, err := soul.Bootstrap(soul.BootstrapParams{
		Name:        in.Name,
		Description: in.Description,
		Accent:      in.Accent,
		AccentAlt:   in.AccentAlt,
		AccentWarm:  in.AccentWarm,
		HeadingFont: in.HeadingFont,
		BodyFont:    in.BodyFont,
		MonoFont:    in.MonoFont,
		Palette:     palette,
		DarkPalette: darkPalette,
	})
	if err != nil {
		return tool.Result[contracts.BootstrapSoulOutput]{}, err
	}
	if err := h.deps.Souls.Put(bootstrapped); err != nil {
		return tool.Result[contracts.BootstrapSoulOutput]{}, err
	}
	tokens := flattenTokens(bootstrapped)
	out := contracts.BootstrapSoulOutput{SoulID: bootstrapped.ID, Name: bootstrapped.Name, Status: contracts.SoulStatus(bootstrapped.Status), TokenCount: len(tokens)}
	return tool.Result[contracts.BootstrapSoulOutput]{Text: fmt.Sprintf("Bootstrapped soul %q (%s).", bootstrapped.Name, bootstrapped.ID), Structured: out}, nil
}

func (h *handlers) refineSoul(_ context.Context, in contracts.RefineSoulInput) (tool.Result[contracts.RefineSoulOutput], error) {
	stored, ok := h.deps.Souls.Get(in.SoulID)
	if !ok {
		return tool.Result[contracts.RefineSoulOutput]{}, fmt.Errorf("soul %q not found", in.SoulID)
	}
	overrides := make([]soul.TokenOverride, 0, len(in.Overrides))
	changed := make([]string, 0, len(in.Overrides))
	for _, override := range in.Overrides {
		overrides = append(overrides, soul.TokenOverride{Category: override.Category, Token: override.Token, Value: override.Value})
		changed = append(changed, strings.TrimSpace(override.Category)+"."+strings.TrimSpace(override.Token))
	}
	refined, err := soul.Refine(stored, overrides)
	if err != nil {
		return tool.Result[contracts.RefineSoulOutput]{}, err
	}
	if err := h.deps.Souls.Put(refined); err != nil {
		return tool.Result[contracts.RefineSoulOutput]{}, err
	}
	tokens := flattenTokens(refined)
	out := contracts.RefineSoulOutput{SoulID: refined.ID, Changed: changed, TokenCount: len(tokens)}
	return tool.Result[contracts.RefineSoulOutput]{Text: fmt.Sprintf("Refined soul %q with %d override(s).", refined.ID, len(changed)), Structured: out}, nil
}

func (h *handlers) listSouls(_ context.Context, in contracts.ListSoulsInput) (tool.Result[contracts.ListSoulsOutput], error) {
	stored := h.deps.Souls.List()
	wantStatus := strings.TrimSpace(string(in.Status))
	out := contracts.ListSoulsOutput{Souls: make([]contracts.SoulSummary, 0, len(stored))}
	for _, item := range stored {
		if wantStatus != "" && item.Status != wantStatus {
			continue
		}
		out.Souls = append(out.Souls, contracts.SoulSummary{SoulID: item.ID, Name: item.Name, Status: contracts.SoulStatus(item.Status), TokenCount: len(flattenTokens(item))})
	}
	return tool.Result[contracts.ListSoulsOutput]{Text: agentText(fmt.Sprintf("Found %d soul(s). Pass a soulId to create_deck (or omit for the default):", len(out.Souls)), out.Souls), Structured: out}, nil
}

func (h *handlers) getSoul(_ context.Context, in contracts.GetSoulInput) (tool.Result[contracts.GetSoulOutput], error) {
	stored, ok := h.deps.Souls.Get(in.SoulID)
	if !ok {
		return tool.Result[contracts.GetSoulOutput]{}, fmt.Errorf("soul %q not found", in.SoulID)
	}
	out := contracts.GetSoulOutput{SoulID: stored.ID, Name: stored.Name, Status: contracts.SoulStatus(stored.Status), Description: stored.Description, Tokens: flattenTokens(stored)}
	if in.IncludeStyleGuide {
		out.StyleGuide = &contracts.SoulStyleGuide{NorthStar: stored.StyleGuide.NorthStar, Do: slices.Clone(stored.StyleGuide.Do), Dont: slices.Clone(stored.StyleGuide.Dont)}
	}
	return tool.Result[contracts.GetSoulOutput]{Text: agentText(fmt.Sprintf("Soul %q:", stored.ID), out), Structured: out}, nil
}

func (h *handlers) getDesignTokens(_ context.Context, in contracts.GetDesignTokensInput) (tool.Result[contracts.GetDesignTokensOutput], error) {
	stored, ok := h.deps.Souls.Get(in.SoulID)
	if !ok {
		return tool.Result[contracts.GetDesignTokensOutput]{}, fmt.Errorf("soul %q not found", in.SoulID)
	}
	out := contracts.GetDesignTokensOutput{Tokens: flattenTokens(stored)}
	return tool.Result[contracts.GetDesignTokensOutput]{Text: agentText(fmt.Sprintf("%d design token(s) for soul %q:", len(out.Tokens), stored.ID), out.Tokens), Structured: out}, nil
}

func flattenTokens(s *soul.Soul) []contracts.TokenEntry {
	if s == nil || s.Theme == nil {
		return nil
	}
	entries := make([]contracts.TokenEntry, 0, len(s.Theme.Colors.Surfaces)+len(s.Theme.Colors.Text)+len(s.Theme.Typography)+len(s.Theme.Spacing)+len(s.Theme.Radii)+len(s.Theme.Elevations)+len(s.Extensions))
	for _, item := range []struct {
		name  string
		value string
		layer contracts.TokenLayer
	}{
		{"canvas", string(s.Theme.Colors.Surfaces[pptx.ColorCanvas]), contracts.TokenLayerSurface},
		{"surface", string(s.Theme.Colors.Surfaces[pptx.ColorSurface]), contracts.TokenLayerSurface},
		{"surfaceAlt", string(s.Theme.Colors.Surfaces[pptx.ColorSurfaceAlt]), contracts.TokenLayerSurface},
		{"accent", string(s.Theme.Colors.Surfaces[pptx.ColorAccent]), contracts.TokenLayerSurface},
		{"accentAlt", string(s.Theme.Colors.Surfaces[pptx.ColorAccentAlt]), contracts.TokenLayerSurface},
		{"accentWarm", string(s.Theme.Colors.Surfaces[pptx.ColorAccentWarm]), contracts.TokenLayerSurface},
		{"success", string(s.Theme.Colors.Surfaces[pptx.ColorSuccess]), contracts.TokenLayerSurface},
		{"warning", string(s.Theme.Colors.Surfaces[pptx.ColorWarning]), contracts.TokenLayerSurface},
		{"error", string(s.Theme.Colors.Surfaces[pptx.ColorError]), contracts.TokenLayerSurface},
		{"info", string(s.Theme.Colors.Surfaces[pptx.ColorInfo]), contracts.TokenLayerSurface},
		{"primary", string(s.Theme.Colors.Text[pptx.TextPrimary]), contracts.TokenLayerText},
		{"secondary", string(s.Theme.Colors.Text[pptx.TextSecondary]), contracts.TokenLayerText},
		{"tertiary", string(s.Theme.Colors.Text[pptx.TextTertiary]), contracts.TokenLayerText},
		{"inverse", string(s.Theme.Colors.Text[pptx.TextInverse]), contracts.TokenLayerText},
		{"muted", string(s.Theme.Colors.Text[pptx.TextMuted]), contracts.TokenLayerText},
		{"accent", string(s.Theme.Colors.Text[pptx.TextAccent]), contracts.TokenLayerText},
		{"accentAlt", string(s.Theme.Colors.Text[pptx.TextAccentAlt]), contracts.TokenLayerText},
		{"success", string(s.Theme.Colors.Text[pptx.TextSuccess]), contracts.TokenLayerText},
		{"warning", string(s.Theme.Colors.Text[pptx.TextWarning]), contracts.TokenLayerText},
		{"error", string(s.Theme.Colors.Text[pptx.TextError]), contracts.TokenLayerText},
		{"display", formatFontSpec(s.Theme.Typography[pptx.TypeDisplay]), contracts.TokenLayerTypography},
		{"h1", formatFontSpec(s.Theme.Typography[pptx.TypeH1]), contracts.TokenLayerTypography},
		{"h2", formatFontSpec(s.Theme.Typography[pptx.TypeH2]), contracts.TokenLayerTypography},
		{"h3", formatFontSpec(s.Theme.Typography[pptx.TypeH3]), contracts.TokenLayerTypography},
		{"h4", formatFontSpec(s.Theme.Typography[pptx.TypeH4]), contracts.TokenLayerTypography},
		{"h5", formatFontSpec(s.Theme.Typography[pptx.TypeH5]), contracts.TokenLayerTypography},
		{"body", formatFontSpec(s.Theme.Typography[pptx.TypeBody]), contracts.TokenLayerTypography},
		{"bodySmall", formatFontSpec(s.Theme.Typography[pptx.TypeBodySmall]), contracts.TokenLayerTypography},
		{"caption", formatFontSpec(s.Theme.Typography[pptx.TypeCaption]), contracts.TokenLayerTypography},
		{"mono", formatFontSpec(s.Theme.Typography[pptx.TypeMono]), contracts.TokenLayerTypography},
		{"code", formatFontSpec(s.Theme.Typography[pptx.TypeCode]), contracts.TokenLayerTypography},
		{"xs", formatPt(s.Theme.Spacing[pptx.SpaceXS]), contracts.TokenLayerSpacing},
		{"sm", formatPt(s.Theme.Spacing[pptx.SpaceSM]), contracts.TokenLayerSpacing},
		{"md", formatPt(s.Theme.Spacing[pptx.SpaceMD]), contracts.TokenLayerSpacing},
		{"lg", formatPt(s.Theme.Spacing[pptx.SpaceLG]), contracts.TokenLayerSpacing},
		{"xl", formatPt(s.Theme.Spacing[pptx.SpaceXL]), contracts.TokenLayerSpacing},
		{"2xl", formatPt(s.Theme.Spacing[pptx.Space2XL]), contracts.TokenLayerSpacing},
		{"none", formatPt(s.Theme.Radii[pptx.RadiusNone]), contracts.TokenLayerRadius},
		{"sm", formatPt(s.Theme.Radii[pptx.RadiusSM]), contracts.TokenLayerRadius},
		{"md", formatPt(s.Theme.Radii[pptx.RadiusMD]), contracts.TokenLayerRadius},
		{"lg", formatPt(s.Theme.Radii[pptx.RadiusLG]), contracts.TokenLayerRadius},
		{"full", formatPt(s.Theme.Radii[pptx.RadiusFull]), contracts.TokenLayerRadius},
		{"flat", formatElevation(s.Theme.Elevations[pptx.ElevationFlat]), contracts.TokenLayerElevation},
		{"raised", formatElevation(s.Theme.Elevations[pptx.ElevationRaised]), contracts.TokenLayerElevation},
		{"elevated", formatElevation(s.Theme.Elevations[pptx.ElevationElevated]), contracts.TokenLayerElevation},
	} {
		entries = append(entries, contracts.TokenEntry{Name: item.name, Value: item.value, Layer: item.layer})
	}
	for _, name := range sortedKeys(s.Extensions) {
		entries = append(entries, contracts.TokenEntry{Name: name, Value: s.Extensions[name], Layer: contracts.TokenLayerExtension})
	}
	return entries
}

func formatFontSpec(spec pptx.FontSpec) string {
	return fmt.Sprintf("family=%s,size=%g,weight=%d,italic=%t", spec.Family, spec.Size, spec.Weight, spec.Italic)
}

func formatPt(value pptx.EMU) string {
	return fmt.Sprintf("%v", value)
}

func formatElevation(value pptx.Elevation) string {
	return fmt.Sprintf("blur=%v,offsetX=%v,offsetY=%v,color=%s,alpha=%d", value.Blur, value.OffsetX, value.OffsetY, value.Color, value.Alpha)
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
