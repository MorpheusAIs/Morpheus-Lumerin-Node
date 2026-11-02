package chat

import (
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/style"
)

func (m model) View() string {
	if m.mode == Chat {
		var b strings.Builder

		b.WriteString(m.viewport.View())
		b.WriteString("\n\n")
		b.WriteString(m.textarea.View())
		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(style.Error.Render(m.err.Error()) + "\n\n")
		}

		b.WriteString(style.Help.Render(common.HelpTextProTip))
		b.WriteString("\n\n")
		b.WriteString(style.Help.Render(common.HelpText))

		return b.String()
	}

	// The picker is where a session gets opened, so the two outcomes of that
	// have to be visible here too. Previously neither was: an open in flight
	// drew nothing at all, and a failed one set m.err where only the Chat
	// branch above renders it, leaving an untouched list and no explanation.
	//
	// Both replace the list rather than appending below it. The list is sized
	// to fill the window, so anything added underneath pushes the terminal
	// into scrolling.
	w := m.messageWidth()

	if m.opening {
		return docStyle.Render(
			m.spinner.View() + "Opening a session for " + m.openingModel + "\n\n" +
				style.Help.Copy().Width(w).Render(
					"This settles a transaction on chain and can take a while.\nctrl+c to give up.") +
				"\n",
		)
	}

	if m.err != nil {
		return docStyle.Render(
			style.Error.Copy().Width(w).Render("Could not open a session: "+m.err.Error()) + "\n\n" +
				style.Help.Copy().Width(w).Render("Press any key to go back to the model list.") +
				"\n",
		)
	}

	return docStyle.Render(m.list.View())
}

// messageWidth is the text width available inside docStyle's margins.
//
// It has to be set explicitly: lipgloss does not wrap at all without a width,
// so a long message is cut off at the right edge of the screen instead. A
// router error carrying a revert reason runs well past 120 characters, and the
// reason is at the end -- precisely the part that was being lost.
func (m model) messageWidth() int {
	h, _ := docStyle.GetFrameSize()
	if w := m.width - h; w >= 40 {
		return w
	}
	return 40
}
