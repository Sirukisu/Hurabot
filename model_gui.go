package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
)

var GuildSelected int
var GuildSelectedCPos int
var GuildSelectedOPos int

func DiscordChannelSelectionGUI() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("guilds", gocui.KeyEnter, gocui.ModNone, selectGuild); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("channels", gocui.KeyEnter, gocui.ModNone, closeGuild); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("channels", gocui.KeySpace, gocui.ModNone, changeChannelEnabled); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("guilds", int(float32(maxX)*0.1), 0, int(float32(maxX)*0.9), int(float32(maxY)*0.8)); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
		v.Title = "Select what channels to include"
		for i := range DiscordGuilds {
			fmt.Fprintln(v, DiscordGuilds[i].Name)
		}
		if _, err := g.SetCurrentView("guilds"); err != nil {
			return err
		}
	}
	return nil
}
func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func cursorDown(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if ((oy + cy) + 1) < len(DiscordGuilds) {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func cursorUp(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func selectGuild(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()

	GuildSelected = oy + cy
	GuildSelectedCPos = cy
	GuildSelectedOPos = oy
	guildInfo := &DiscordGuilds[GuildSelected]

	maxX, maxY := g.Size()
	if v, err := g.SetView("channels", int(float32(maxX)*0.2), int(float32(maxY)*0.2), int(float32(maxX)*0.8), int(float32(maxY)*0.8)); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Guild " + guildInfo.Name + " channels"
		v.Highlight = true
		v.SelBgColor = gocui.ColorRed
		v.SelFgColor = gocui.ColorBlack

		checkbox := "[ ]"

		for _, channel := range guildInfo.Channels {
			if channel.Enabled == true {
				checkbox = "[X]"
			} else {
				checkbox = "[ ]"
			}
			fmt.Fprintln(v, checkbox+" "+channel.Name)

		}
		if _, err := g.SetCurrentView("channels"); err != nil {
			return err
		}
	}
	return nil
}

func closeGuild(g *gocui.Gui, _ *gocui.View) error {
	if err := g.DeleteView("channels"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("guilds"); err != nil {
		return err
	}
	return nil
}

func changeChannelEnabled(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()

	guildInfo := &DiscordGuilds[GuildSelected]

	channelInfo := &guildInfo.Channels[cy+oy]

	if channelInfo.Enabled == false {
		channelInfo.Enabled = true
	} else {
		channelInfo.Enabled = false
	}

	g.Update(func(g *gocui.Gui) error {
		closeGuild(g, v)
		view, err := g.View("guilds")
		if err != nil {
			return err
		}
		view.SetOrigin(0, GuildSelectedOPos)
		view.SetCursor(0, GuildSelectedCPos)

		selectGuild(g, view)
		view2, err := g.View("channels")
		if err != nil {
			return err
		}
		view2.SetOrigin(0, oy)
		view2.SetCursor(0, cy)

		return nil
	})
	return nil
}
