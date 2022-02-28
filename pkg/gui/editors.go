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
	branch, err := osCommand.RunCommandWithOutput(getBranch)
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
	log, err := osCommand.RunCommandWithOutput(gitLogCommand)
	if err != nil {
		fmt.Fprintln(gui.Views.Extras, err)
	}
	fmt.Fprintln(gui.Views.Extras, "\n"+gitLogCommand)
	return log
}

// we've just copy+pasted the editor from gocui to here so that we can also re-
// render the commit message length on each keypress
func (gui *Gui) commitMessageEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	newlineKey, ok := gui.getKey(gui.Config.GetUserConfig().Keybinding.Universal.AppendNewline).(gocui.Key)
	if !ok {
		newlineKey = gocui.KeyAltEnter
	}

	matched := true
	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyCtrlD || key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyArrowDown:
		if num >= 0 {
			v.SetEditorContent(gui.gitLog())
		}
		num--
		// v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp:
		if num <= 100 {
			v.SetEditorContent(gui.gitLog())
		}
		num++
		// v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	case key == newlineKey:
		v.EditNewLine()
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyCtrlU:
		v.EditDeleteToStartOfLine()
	case key == gocui.KeyTab:
		v.SetEditorContent(gui.getBranch())
	case key == gocui.KeyCtrlA:
		v.EditGotoToStartOfLine()
	case key == gocui.KeyCtrlE:
		v.EditGotoToEndOfLine()

		// TODO: see if we need all three of these conditions: maybe the final one is sufficient
	case ch != 0 && mod == 0 && unicode.IsPrint(ch):
		v.EditWrite(ch)
	default:
		matched = false
	}

	gui.RenderCommitLength()

	return matched
}

func (gui *Gui) defaultEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	matched := true
	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyCtrlD || key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyArrowDown:
		v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyCtrlU:
		v.EditDeleteToStartOfLine()
	case key == gocui.KeyCtrlA:
		v.EditGotoToStartOfLine()
	case key == gocui.KeyCtrlE:
		v.EditGotoToEndOfLine()

		// TODO: see if we need all three of these conditions: maybe the final one is sufficient
	case ch != 0 && mod == 0 && unicode.IsPrint(ch):
		v.EditWrite(ch)
	default:
		matched = false
	}

	if gui.findSuggestions != nil {
		input := v.Buffer()
		suggestions := gui.findSuggestions(input)
		gui.setSuggestions(suggestions)
	}

	return matched
}
