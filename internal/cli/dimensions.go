package cli

// Dimensions определяет ограничения размеров для responsive layout.
type Dimensions struct {
	MinListWidth int // Минимальная ширина списка (в символах)
	MaxListWidth int // Максимальная ширина списка (в символах)
	MinHeight    int // Минимальная высота окна
	PaddingLeft  int // Padding слева (по плану 20px)
	PaddingRight int // Padding справа (по плану 20px)
	PaddingTop   int // Padding сверху (по плану 10px ≈ 5 строк)
	PaddingBottom int // Padding снизу (по плану 10px ≈ 5 строк)
}

// DefaultDimensions возвращает значения по умолчанию согласно плану.
func DefaultDimensions() Dimensions {
	return Dimensions{
		MinListWidth:  40,
		MaxListWidth:  120,
		MinHeight:     10,
		PaddingLeft:   20,
		PaddingRight:  20,
		PaddingTop:    5,
		PaddingBottom: 5,
	}
}

// EffectiveWidth вычисляет эффективную ширину списка с учетом padding и ограничений.
func (d *Dimensions) EffectiveWidth(windowWidth int) int {
	available := windowWidth - d.PaddingLeft - d.PaddingRight
	if available < d.MinListWidth {
		return d.MinListWidth
	}
	if available > d.MaxListWidth && d.MaxListWidth > 0 {
		return d.MaxListWidth
	}
	return available
}

// EffectiveHeight вычисляет эффективную высоту списка с учетом padding.
func (d *Dimensions) EffectiveHeight(windowHeight int) int {
	available := windowHeight - d.PaddingTop - d.PaddingBottom
	if available < 3 {
		return 3 // Минимальная высота — 3 строки
	}
	return available
}

// ClampSize применяет ограничения к заданной ширине и высоте.
func (d *Dimensions) ClampSize(width, height int) (int, int) {
	effectiveW := d.EffectiveWidth(width)
	effectiveH := d.EffectiveHeight(height)
	return effectiveW, effectiveH
}

