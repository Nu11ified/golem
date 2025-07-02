//go:build js && wasm

package css

import (
	"fmt"
	"strings"
	"syscall/js"
)

// Style represents a CSS style declaration
type Style struct {
	Property string
	Value    interface{}
}

// StyleSheet manages CSS styles
type StyleSheet struct {
	rules        map[string][]Style
	keyframes    map[string][]Keyframe
	vars         map[string]string
	mediaQueries map[string][]Rule
}

// Rule represents a CSS rule
type Rule struct {
	Selector string
	Styles   []Style
}

// Keyframe represents a CSS keyframe
type Keyframe struct {
	Offset string // "0%", "50%", "100%", "from", "to"
	Styles []Style
}

// NewStyleSheet creates a new stylesheet
func NewStyleSheet() *StyleSheet {
	return &StyleSheet{
		rules:        make(map[string][]Style),
		keyframes:    make(map[string][]Keyframe),
		vars:         make(map[string]string),
		mediaQueries: make(map[string][]Rule),
	}
}

// CSS property builder functions
func Property(name string, value interface{}) Style {
	return Style{Property: name, Value: value}
}

// Layout properties
func Display(value string) Style        { return Property("display", value) }
func Position(value string) Style       { return Property("position", value) }
func Top(value interface{}) Style       { return Property("top", value) }
func Right(value interface{}) Style     { return Property("right", value) }
func Bottom(value interface{}) Style    { return Property("bottom", value) }
func Left(value interface{}) Style      { return Property("left", value) }
func Width(value interface{}) Style     { return Property("width", value) }
func Height(value interface{}) Style    { return Property("height", value) }
func MinWidth(value interface{}) Style  { return Property("min-width", value) }
func MinHeight(value interface{}) Style { return Property("min-height", value) }
func MaxWidth(value interface{}) Style  { return Property("max-width", value) }
func MaxHeight(value interface{}) Style { return Property("max-height", value) }

// Flexbox properties
func FlexDirection(value string) Style   { return Property("flex-direction", value) }
func FlexWrap(value string) Style        { return Property("flex-wrap", value) }
func JustifyContent(value string) Style  { return Property("justify-content", value) }
func AlignItems(value string) Style      { return Property("align-items", value) }
func AlignContent(value string) Style    { return Property("align-content", value) }
func Flex(value interface{}) Style       { return Property("flex", value) }
func FlexGrow(value interface{}) Style   { return Property("flex-grow", value) }
func FlexShrink(value interface{}) Style { return Property("flex-shrink", value) }
func FlexBasis(value interface{}) Style  { return Property("flex-basis", value) }
func AlignSelf(value string) Style       { return Property("align-self", value) }

// Grid properties
func GridTemplateColumns(value string) Style { return Property("grid-template-columns", value) }
func GridTemplateRows(value string) Style    { return Property("grid-template-rows", value) }
func GridColumn(value string) Style          { return Property("grid-column", value) }
func GridRow(value string) Style             { return Property("grid-row", value) }
func GridGap(value interface{}) Style        { return Property("grid-gap", value) }
func GridColumnGap(value interface{}) Style  { return Property("grid-column-gap", value) }
func GridRowGap(value interface{}) Style     { return Property("grid-row-gap", value) }

// Typography properties
func FontFamily(value string) Style         { return Property("font-family", value) }
func FontSize(value interface{}) Style      { return Property("font-size", value) }
func FontWeight(value interface{}) Style    { return Property("font-weight", value) }
func FontStyle(value string) Style          { return Property("font-style", value) }
func LineHeight(value interface{}) Style    { return Property("line-height", value) }
func TextAlign(value string) Style          { return Property("text-align", value) }
func TextDecoration(value string) Style     { return Property("text-decoration", value) }
func TextTransform(value string) Style      { return Property("text-transform", value) }
func LetterSpacing(value interface{}) Style { return Property("letter-spacing", value) }
func WordSpacing(value interface{}) Style   { return Property("word-spacing", value) }

// Color and background properties
func Color(value string) Style              { return Property("color", value) }
func BackgroundColor(value string) Style    { return Property("background-color", value) }
func Background(value string) Style         { return Property("background", value) }
func BackgroundImage(value string) Style    { return Property("background-image", value) }
func BackgroundSize(value string) Style     { return Property("background-size", value) }
func BackgroundPosition(value string) Style { return Property("background-position", value) }
func BackgroundRepeat(value string) Style   { return Property("background-repeat", value) }

// Border properties
func Border(value string) Style            { return Property("border", value) }
func BorderTop(value string) Style         { return Property("border-top", value) }
func BorderRight(value string) Style       { return Property("border-right", value) }
func BorderBottom(value string) Style      { return Property("border-bottom", value) }
func BorderLeft(value string) Style        { return Property("border-left", value) }
func BorderWidth(value interface{}) Style  { return Property("border-width", value) }
func BorderStyle(value string) Style       { return Property("border-style", value) }
func BorderColor(value string) Style       { return Property("border-color", value) }
func BorderRadius(value interface{}) Style { return Property("border-radius", value) }

// Spacing properties
func Margin(value interface{}) Style        { return Property("margin", value) }
func MarginTop(value interface{}) Style     { return Property("margin-top", value) }
func MarginRight(value interface{}) Style   { return Property("margin-right", value) }
func MarginBottom(value interface{}) Style  { return Property("margin-bottom", value) }
func MarginLeft(value interface{}) Style    { return Property("margin-left", value) }
func Padding(value interface{}) Style       { return Property("padding", value) }
func PaddingTop(value interface{}) Style    { return Property("padding-top", value) }
func PaddingRight(value interface{}) Style  { return Property("padding-right", value) }
func PaddingBottom(value interface{}) Style { return Property("padding-bottom", value) }
func PaddingLeft(value interface{}) Style   { return Property("padding-left", value) }

// Visual effects
func Opacity(value interface{}) Style { return Property("opacity", value) }
func Visibility(value string) Style   { return Property("visibility", value) }
func Overflow(value string) Style     { return Property("overflow", value) }
func OverflowX(value string) Style    { return Property("overflow-x", value) }
func OverflowY(value string) Style    { return Property("overflow-y", value) }
func ZIndex(value interface{}) Style  { return Property("z-index", value) }
func BoxShadow(value string) Style    { return Property("box-shadow", value) }
func TextShadow(value string) Style   { return Property("text-shadow", value) }

// Transform and animation
func Transform(value string) Style       { return Property("transform", value) }
func TransformOrigin(value string) Style { return Property("transform-origin", value) }
func Transition(value string) Style      { return Property("transition", value) }
func Animation(value string) Style       { return Property("animation", value) }

// Cursor and interaction
func Cursor(value string) Style        { return Property("cursor", value) }
func PointerEvents(value string) Style { return Property("pointer-events", value) }
func UserSelect(value string) Style    { return Property("user-select", value) }

// CSS-in-Go composition helpers
type StyleBuilder struct {
	styles []Style
}

// NewStyleBuilder creates a new style builder
func NewStyleBuilder() *StyleBuilder {
	return &StyleBuilder{styles: make([]Style, 0)}
}

// Add adds styles to the builder
func (sb *StyleBuilder) Add(styles ...Style) *StyleBuilder {
	sb.styles = append(sb.styles, styles...)
	return sb
}

// When conditionally adds styles
func (sb *StyleBuilder) When(condition bool, styles ...Style) *StyleBuilder {
	if condition {
		sb.styles = append(sb.styles, styles...)
	}
	return sb
}

// Build returns the final styles
func (sb *StyleBuilder) Build() []Style {
	return sb.styles
}

// Responsive design helpers
type Breakpoint struct {
	Name  string
	Query string
}

var (
	Mobile  = Breakpoint{"mobile", "max-width: 768px"}
	Tablet  = Breakpoint{"tablet", "min-width: 769px and max-width: 1024px"}
	Desktop = Breakpoint{"desktop", "min-width: 1025px"}
)

// MediaQuery creates a media query rule
func (ss *StyleSheet) MediaQuery(breakpoint Breakpoint, rules ...Rule) {
	if ss.mediaQueries[breakpoint.Query] == nil {
		ss.mediaQueries[breakpoint.Query] = make([]Rule, 0)
	}
	ss.mediaQueries[breakpoint.Query] = append(ss.mediaQueries[breakpoint.Query], rules...)
}

// CSS Variables
func (ss *StyleSheet) SetVariable(name, value string) {
	ss.vars[name] = value
}

func Var(name string) string {
	return fmt.Sprintf("var(--%s)", name)
}

// Animation helpers
func (ss *StyleSheet) AddKeyframes(name string, keyframes []Keyframe) {
	ss.keyframes[name] = keyframes
}

func KeyframeFrom(styles ...Style) Keyframe {
	return Keyframe{Offset: "from", Styles: styles}
}

func KeyframeTo(styles ...Style) Keyframe {
	return Keyframe{Offset: "to", Styles: styles}
}

func KeyframeAt(percentage string, styles ...Style) Keyframe {
	return Keyframe{Offset: percentage, Styles: styles}
}

// Style application
func (ss *StyleSheet) AddRule(selector string, styles ...Style) {
	ss.rules[selector] = styles
}

// Generate CSS string
func (ss *StyleSheet) String() string {
	var css strings.Builder

	// CSS Variables
	if len(ss.vars) > 0 {
		css.WriteString(":root {\n")
		for name, value := range ss.vars {
			css.WriteString(fmt.Sprintf("  --%s: %s;\n", name, value))
		}
		css.WriteString("}\n\n")
	}

	// Regular rules
	for selector, styles := range ss.rules {
		css.WriteString(fmt.Sprintf("%s {\n", selector))
		for _, style := range styles {
			css.WriteString(fmt.Sprintf("  %s: %v;\n", style.Property, style.Value))
		}
		css.WriteString("}\n\n")
	}

	// Keyframes
	for name, keyframes := range ss.keyframes {
		css.WriteString(fmt.Sprintf("@keyframes %s {\n", name))
		for _, kf := range keyframes {
			css.WriteString(fmt.Sprintf("  %s {\n", kf.Offset))
			for _, style := range kf.Styles {
				css.WriteString(fmt.Sprintf("    %s: %v;\n", style.Property, style.Value))
			}
			css.WriteString("  }\n")
		}
		css.WriteString("}\n\n")
	}

	// Media queries
	for query, rules := range ss.mediaQueries {
		css.WriteString(fmt.Sprintf("@media (%s) {\n", query))
		for _, rule := range rules {
			css.WriteString(fmt.Sprintf("  %s {\n", rule.Selector))
			for _, style := range rule.Styles {
				css.WriteString(fmt.Sprintf("    %s: %v;\n", style.Property, style.Value))
			}
			css.WriteString("  }\n")
		}
		css.WriteString("}\n\n")
	}

	return css.String()
}

// Inject styles into the document
func (ss *StyleSheet) Inject() {
	doc := js.Global().Get("document")
	head := doc.Get("head")

	// Create style element
	styleEl := doc.Call("createElement", "style")
	styleEl.Set("textContent", ss.String())

	// Append to head
	head.Call("appendChild", styleEl)
}

// Pre-built style utilities
type Utilities struct{}

var Utils = &Utilities{}

// Common utility styles
func (u *Utilities) FlexCenter() []Style {
	return []Style{
		Display("flex"),
		JustifyContent("center"),
		AlignItems("center"),
	}
}

func (u *Utilities) FlexColumn() []Style {
	return []Style{
		Display("flex"),
		FlexDirection("column"),
	}
}

func (u *Utilities) FlexRow() []Style {
	return []Style{
		Display("flex"),
		FlexDirection("row"),
	}
}

func (u *Utilities) FullSize() []Style {
	return []Style{
		Width("100%"),
		Height("100%"),
	}
}

func (u *Utilities) CenterText() []Style {
	return []Style{
		TextAlign("center"),
	}
}

func (u *Utilities) Hidden() []Style {
	return []Style{
		Display("none"),
	}
}

func (u *Utilities) Visible() []Style {
	return []Style{
		Display("block"),
	}
}

func (u *Utilities) Card() []Style {
	return []Style{
		BackgroundColor("white"),
		BorderRadius("8px"),
		BoxShadow("0 2px 4px rgba(0,0,0,0.1)"),
		Padding("16px"),
	}
}

func (u *Utilities) Button() []Style {
	return []Style{
		Display("inline-block"),
		Padding("10px 20px"),
		BackgroundColor("#007bff"),
		Color("white"),
		Border("none"),
		BorderRadius("4px"),
		Cursor("pointer"),
		TextDecoration("none"),
		Transition("background-color 0.2s"),
	}
}

// Theme system
type Theme struct {
	Colors      map[string]string
	Fonts       map[string]string
	Spacing     map[string]string
	Breakpoints map[string]string
}

func NewTheme() *Theme {
	return &Theme{
		Colors: map[string]string{
			"primary":   "#007bff",
			"secondary": "#6c757d",
			"success":   "#28a745",
			"danger":    "#dc3545",
			"warning":   "#ffc107",
			"info":      "#17a2b8",
			"light":     "#f8f9fa",
			"dark":      "#343a40",
		},
		Fonts: map[string]string{
			"sans":  "system-ui, -apple-system, sans-serif",
			"serif": "Georgia, serif",
			"mono":  "Menlo, Monaco, monospace",
		},
		Spacing: map[string]string{
			"xs": "4px",
			"sm": "8px",
			"md": "16px",
			"lg": "24px",
			"xl": "32px",
		},
		Breakpoints: map[string]string{
			"sm": "576px",
			"md": "768px",
			"lg": "992px",
			"xl": "1200px",
		},
	}
}

func (t *Theme) Color(name string) string {
	if color, exists := t.Colors[name]; exists {
		return color
	}
	return name
}

func (t *Theme) Font(name string) string {
	if font, exists := t.Fonts[name]; exists {
		return font
	}
	return name
}

func (t *Theme) Space(name string) string {
	if space, exists := t.Spacing[name]; exists {
		return space
	}
	return name
}

// Global theme instance
var DefaultTheme = NewTheme()

// Styled component helpers
type StyledComponent struct {
	BaseStyles []Style
	States     map[string][]Style
}

func NewStyledComponent(baseStyles ...Style) *StyledComponent {
	return &StyledComponent{
		BaseStyles: baseStyles,
		States:     make(map[string][]Style),
	}
}

func (sc *StyledComponent) AddState(state string, styles ...Style) *StyledComponent {
	sc.States[state] = styles
	return sc
}

func (sc *StyledComponent) Hover(styles ...Style) *StyledComponent {
	return sc.AddState("hover", styles...)
}

func (sc *StyledComponent) Focus(styles ...Style) *StyledComponent {
	return sc.AddState("focus", styles...)
}

func (sc *StyledComponent) Active(styles ...Style) *StyledComponent {
	return sc.AddState("active", styles...)
}

func (sc *StyledComponent) GenerateCSS(className string) string {
	var css strings.Builder

	// Base styles
	css.WriteString(fmt.Sprintf(".%s {\n", className))
	for _, style := range sc.BaseStyles {
		css.WriteString(fmt.Sprintf("  %s: %v;\n", style.Property, style.Value))
	}
	css.WriteString("}\n")

	// State styles
	for state, styles := range sc.States {
		css.WriteString(fmt.Sprintf(".%s:%s {\n", className, state))
		for _, style := range styles {
			css.WriteString(fmt.Sprintf("  %s: %v;\n", style.Property, style.Value))
		}
		css.WriteString("}\n")
	}

	return css.String()
}

// CSS class name generation
var classCounter = 0

func GenerateClassName(prefix string) string {
	classCounter++
	return fmt.Sprintf("%s-%d", prefix, classCounter)
}

// Runtime style injection
func InjectStyles(css string) {
	doc := js.Global().Get("document")
	head := doc.Call("querySelector", "head")
	if head.IsNull() {
		fmt.Println("Could not find head element to inject styles")
		return
	}

	styleElement := doc.Call("createElement", "style")
	styleElement.Set("innerHTML", css)
	head.Call("appendChild", styleElement)
}
