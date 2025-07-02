//go:build !js || !wasm

package css

import "fmt"

// Stub implementations for non-WASM builds
type Style struct {
	Property string
	Value    interface{}
}

type StyleSheet struct {
	rules        map[string][]Style
	keyframes    map[string][]Keyframe
	vars         map[string]string
	mediaQueries map[string][]Rule
}

type Rule struct {
	Selector string
	Styles   []Style
}

type Keyframe struct {
	Offset string
	Styles []Style
}

type StyleBuilder struct {
	styles []Style
}

type Breakpoint struct {
	Name  string
	Query string
}

type Theme struct {
	Colors      map[string]string
	Fonts       map[string]string
	Spacing     map[string]string
	Breakpoints map[string]string
}

type StyledComponent struct {
	BaseStyles []Style
	States     map[string][]Style
}

type Utilities struct{}

// Stub functions
func NewStyleSheet() *StyleSheet {
	return &StyleSheet{
		rules:        make(map[string][]Style),
		keyframes:    make(map[string][]Keyframe),
		vars:         make(map[string]string),
		mediaQueries: make(map[string][]Rule),
	}
}

func Property(name string, value interface{}) Style {
	return Style{Property: name, Value: value}
}

// All the CSS property functions
func Display(value string) Style           { return Property("display", value) }
func Position(value string) Style          { return Property("position", value) }
func Width(value interface{}) Style        { return Property("width", value) }
func Height(value interface{}) Style       { return Property("height", value) }
func Color(value string) Style             { return Property("color", value) }
func BackgroundColor(value string) Style   { return Property("background-color", value) }
func Margin(value interface{}) Style       { return Property("margin", value) }
func Padding(value interface{}) Style      { return Property("padding", value) }
func FontSize(value interface{}) Style     { return Property("font-size", value) }
func TextAlign(value string) Style         { return Property("text-align", value) }
func Border(value string) Style            { return Property("border", value) }
func BorderRadius(value interface{}) Style { return Property("border-radius", value) }

// Additional stubs for other functions
func Top(value interface{}) Style            { return Property("top", value) }
func Right(value interface{}) Style          { return Property("right", value) }
func Bottom(value interface{}) Style         { return Property("bottom", value) }
func Left(value interface{}) Style           { return Property("left", value) }
func MinWidth(value interface{}) Style       { return Property("min-width", value) }
func MinHeight(value interface{}) Style      { return Property("min-height", value) }
func MaxWidth(value interface{}) Style       { return Property("max-width", value) }
func MaxHeight(value interface{}) Style      { return Property("max-height", value) }
func FlexDirection(value string) Style       { return Property("flex-direction", value) }
func FlexWrap(value string) Style            { return Property("flex-wrap", value) }
func JustifyContent(value string) Style      { return Property("justify-content", value) }
func AlignItems(value string) Style          { return Property("align-items", value) }
func AlignContent(value string) Style        { return Property("align-content", value) }
func Flex(value interface{}) Style           { return Property("flex", value) }
func FlexGrow(value interface{}) Style       { return Property("flex-grow", value) }
func FlexShrink(value interface{}) Style     { return Property("flex-shrink", value) }
func FlexBasis(value interface{}) Style      { return Property("flex-basis", value) }
func AlignSelf(value string) Style           { return Property("align-self", value) }
func GridTemplateColumns(value string) Style { return Property("grid-template-columns", value) }
func GridTemplateRows(value string) Style    { return Property("grid-template-rows", value) }
func GridColumn(value string) Style          { return Property("grid-column", value) }
func GridRow(value string) Style             { return Property("grid-row", value) }
func GridGap(value interface{}) Style        { return Property("grid-gap", value) }
func GridColumnGap(value interface{}) Style  { return Property("grid-column-gap", value) }
func GridRowGap(value interface{}) Style     { return Property("grid-row-gap", value) }
func FontFamily(value string) Style          { return Property("font-family", value) }
func FontWeight(value interface{}) Style     { return Property("font-weight", value) }
func FontStyle(value string) Style           { return Property("font-style", value) }
func LineHeight(value interface{}) Style     { return Property("line-height", value) }
func TextDecoration(value string) Style      { return Property("text-decoration", value) }
func TextTransform(value string) Style       { return Property("text-transform", value) }
func LetterSpacing(value interface{}) Style  { return Property("letter-spacing", value) }
func WordSpacing(value interface{}) Style    { return Property("word-spacing", value) }
func Background(value string) Style          { return Property("background", value) }
func BackgroundImage(value string) Style     { return Property("background-image", value) }
func BackgroundSize(value string) Style      { return Property("background-size", value) }
func BackgroundPosition(value string) Style  { return Property("background-position", value) }
func BackgroundRepeat(value string) Style    { return Property("background-repeat", value) }
func BorderTop(value string) Style           { return Property("border-top", value) }
func BorderRight(value string) Style         { return Property("border-right", value) }
func BorderBottom(value string) Style        { return Property("border-bottom", value) }
func BorderLeft(value string) Style          { return Property("border-left", value) }
func BorderWidth(value interface{}) Style    { return Property("border-width", value) }
func BorderStyle(value string) Style         { return Property("border-style", value) }
func BorderColor(value string) Style         { return Property("border-color", value) }
func MarginTop(value interface{}) Style      { return Property("margin-top", value) }
func MarginRight(value interface{}) Style    { return Property("margin-right", value) }
func MarginBottom(value interface{}) Style   { return Property("margin-bottom", value) }
func MarginLeft(value interface{}) Style     { return Property("margin-left", value) }
func PaddingTop(value interface{}) Style     { return Property("padding-top", value) }
func PaddingRight(value interface{}) Style   { return Property("padding-right", value) }
func PaddingBottom(value interface{}) Style  { return Property("padding-bottom", value) }
func PaddingLeft(value interface{}) Style    { return Property("padding-left", value) }
func Opacity(value interface{}) Style        { return Property("opacity", value) }
func Visibility(value string) Style          { return Property("visibility", value) }
func Overflow(value string) Style            { return Property("overflow", value) }
func OverflowX(value string) Style           { return Property("overflow-x", value) }
func OverflowY(value string) Style           { return Property("overflow-y", value) }
func ZIndex(value interface{}) Style         { return Property("z-index", value) }
func BoxShadow(value string) Style           { return Property("box-shadow", value) }
func TextShadow(value string) Style          { return Property("text-shadow", value) }
func Transform(value string) Style           { return Property("transform", value) }
func TransformOrigin(value string) Style     { return Property("transform-origin", value) }
func Transition(value string) Style          { return Property("transition", value) }
func Animation(value string) Style           { return Property("animation", value) }
func Cursor(value string) Style              { return Property("cursor", value) }
func PointerEvents(value string) Style       { return Property("pointer-events", value) }
func UserSelect(value string) Style          { return Property("user-select", value) }

func NewStyleBuilder() *StyleBuilder {
	return &StyleBuilder{styles: make([]Style, 0)}
}

func (sb *StyleBuilder) Add(styles ...Style) *StyleBuilder {
	sb.styles = append(sb.styles, styles...)
	return sb
}

func (sb *StyleBuilder) When(condition bool, styles ...Style) *StyleBuilder {
	if condition {
		sb.styles = append(sb.styles, styles...)
	}
	return sb
}

func (sb *StyleBuilder) Build() []Style {
	return sb.styles
}

var (
	Mobile  = Breakpoint{"mobile", "max-width: 768px"}
	Tablet  = Breakpoint{"tablet", "min-width: 769px and max-width: 1024px"}
	Desktop = Breakpoint{"desktop", "min-width: 1025px"}
)

func (ss *StyleSheet) MediaQuery(breakpoint Breakpoint, rules ...Rule) {
	// No-op for stub
}

func (ss *StyleSheet) SetVariable(name, value string) {
	ss.vars[name] = value
}

func Var(name string) string {
	return fmt.Sprintf("var(--%s)", name)
}

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

func (ss *StyleSheet) AddRule(selector string, styles ...Style) {
	ss.rules[selector] = styles
}

func (ss *StyleSheet) String() string {
	return "/* CSS generation only available in WebAssembly build */"
}

func (ss *StyleSheet) Inject() {
	fmt.Println("CSS injection only available in WebAssembly build")
}

var Utils = &Utilities{}

func (u *Utilities) FlexCenter() []Style { return []Style{} }
func (u *Utilities) FlexColumn() []Style { return []Style{} }
func (u *Utilities) FlexRow() []Style    { return []Style{} }
func (u *Utilities) FullSize() []Style   { return []Style{} }
func (u *Utilities) CenterText() []Style { return []Style{} }
func (u *Utilities) Hidden() []Style     { return []Style{} }
func (u *Utilities) Visible() []Style    { return []Style{} }
func (u *Utilities) Card() []Style       { return []Style{} }
func (u *Utilities) Button() []Style     { return []Style{} }

func NewTheme() *Theme {
	return &Theme{
		Colors:      make(map[string]string),
		Fonts:       make(map[string]string),
		Spacing:     make(map[string]string),
		Breakpoints: make(map[string]string),
	}
}

func (t *Theme) Color(name string) string { return name }
func (t *Theme) Font(name string) string  { return name }
func (t *Theme) Space(name string) string { return name }

var DefaultTheme = NewTheme()

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

func (sc *StyledComponent) Hover(styles ...Style) *StyledComponent  { return sc }
func (sc *StyledComponent) Focus(styles ...Style) *StyledComponent  { return sc }
func (sc *StyledComponent) Active(styles ...Style) *StyledComponent { return sc }

func (sc *StyledComponent) GenerateCSS(className string) string {
	return "/* CSS generation only available in WebAssembly build */"
}

var classCounter = 0

func GenerateClassName(prefix string) string {
	classCounter++
	return fmt.Sprintf("%s-%d", prefix, classCounter)
}

func InjectStyles(css string) {
	fmt.Printf("CSS injection only available in WebAssembly build: %s\n", css)
}
