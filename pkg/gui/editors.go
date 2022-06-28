package gui

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
)

var num int = 0

// 填写 commit 信息的时候，根据业务分支自动增加 commit 前缀
func (gui *Gui) getBranch() string {
	osCommand := oscommands.NewDummyOSCommand()
	getBranch := "git rev-parse --abbrev-ref HEAD"
	fmt.Fprintln(gui.Views.Extras, "\n"+getBranch)
	branch, err := osCommand.Cmd.New(getBranch).RunWithOutput()
	if err != nil {
		fmt.Fprintln(gui.Views.Extras, err)
	}
	// fmt.Fprintln(gui.Views.Extras, style.FgCyan.Sprint(branch))
	flysnowRegexp := regexp.MustCompile(`([A-Z]+-\d+)|(hotfix)`)
	params := flysnowRegexp.FindStringSubmatch(branch)
	if len(params) != 0 {
		var commitPrefix string
		if params[1] == "hotfix" {
			commitPrefix = "FE-0000: "
		} else {
			commitPrefix = params[1] + ": "
		}
		return commitPrefix
	}
	return ""
}

// commit message log
func (gui *Gui) gitLog() string {
	osCommand := oscommands.NewDummyOSCommand()
	gitLogCommand := fmt.Sprintf("git log -n 1 --skip %s --pretty=format:%s", strconv.Itoa(num), "%s")
	log, err := osCommand.Cmd.New(gitLogCommand).RunWithOutput()
	if err != nil {
		fmt.Fprintln(gui.Views.Extras, err)
	}
	fmt.Fprintln(gui.Views.Extras, "\n"+gitLogCommand)
	return log
}

func (gui *Gui) handleEditorKeypress(textArea *gocui.TextArea, key gocui.Key, ch rune, mod gocui.Modifier, allowMultiline bool) bool {
	newlineKey, ok := gui.getKey(gui.c.UserConfig.Keybinding.Universal.AppendNewline).(gocui.Key)
	if !ok {
		newlineKey = gocui.KeyAltEnter
	}

	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		textArea.BackSpaceChar()
	case key == gocui.KeyCtrlD || key == gocui.KeyDelete:
		textArea.DeleteChar()
	case key == gocui.KeyArrowDown:
		if num >= 0 {
			textArea.Clear()
			textArea.TypeString(gui.gitLog())
		}
		num--
		//textArea.MoveCursorDown()
	case key == gocui.KeyArrowUp:
		if num <= 100 {
			textArea.Clear()
			textArea.TypeString(gui.gitLog())
		}
		num++
		// textArea.MoveCursorUp()
	case key == gocui.KeyArrowLeft:
		textArea.MoveCursorLeft()
	case key == gocui.KeyArrowRight:
		textArea.MoveCursorRight()
	case key == newlineKey:
		if allowMultiline {
			textArea.TypeRune('\n')
		} else {
			return false
		}
	case key == gocui.KeySpace:
		textArea.TypeRune(' ')
	case key == gocui.KeyInsert:
		textArea.ToggleOverwrite()
	case key == gocui.KeyCtrlU:
		textArea.DeleteToStartOfLine()
	case key == gocui.KeyCtrlA || key == gocui.KeyHome:
		textArea.GoToStartOfLine()
	case key == gocui.KeyCtrlE || key == gocui.KeyEnd:
		textArea.GoToEndOfLine()
	case key == gocui.KeyTab:
		textArea.Clear()
		textArea.TypeString(gui.getBranch())

		// TODO: see if we need all three of these conditions: maybe the final one is sufficient
	case ch != 0 && mod == 0 && unicode.IsPrint(ch):
		textArea.TypeRune(ch)
	default:
		return false
	}

	return true
}

// we've just copy+pasted the editor from gocui to here so that we can also re-
// render the commit message length on each keypress
func (gui *Gui) commitMessageEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	matched := gui.handleEditorKeypress(v.TextArea, key, ch, mod, true)

	// This function is called again on refresh as part of the general resize popup call,
	// but we need to call it here so that when we go to render the text area it's not
	// considered out of bounds to add a newline, meaning we can avoid unnecessary scrolling.
	err := gui.resizePopupPanel(v, v.TextArea.GetContent())
	if err != nil {
		gui.c.Log.Error(err)
	}
	v.RenderTextArea()
	gui.RenderCommitLength()

	return matched
}

func (gui *Gui) defaultEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	matched := gui.handleEditorKeypress(v.TextArea, key, ch, mod, false)

	v.RenderTextArea()

	if gui.findSuggestions != nil {
		input := v.TextArea.GetContent()
		gui.suggestionsAsyncHandler.Do(func() func() {
			suggestions := gui.findSuggestions(input)
			return func() { gui.setSuggestions(suggestions) }
		})
	}

	return matched
}
