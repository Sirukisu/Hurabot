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
	if err := g.SetKeybinding("guilds", gocui.KeyArrowDown, gocui.ModNone, cursorDownGuilds); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("guilds", gocui.KeyEnter, gocui.ModNone, selectGuild); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("guilds", gocui.KeySpace, gocui.ModNone, changeGuildEnabled); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("channels", gocui.KeyEnter, gocui.ModNone, closeGuild); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("channels", gocui.KeyArrowDown, gocui.ModNone, cursorDownChannels); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("channels", gocui.KeySpace, gocui.ModNone, changeChannelEnabled); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("guilds", gocui.KeyCtrlD, gocui.ModNone, saveName); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("modelName", gocui.KeyEnter, gocui.ModNone, confirmName); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("modelName", gocui.KeyCtrlD, gocui.ModNone, closeConfirm); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("guilds", int(float32(maxX)*0.05), 0, int(float32(maxX)*0.95), int(float32(maxY)*0.8)); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
		v.Title = "Select what channels to include"

		drawGuilds(v)
		if _, err := g.SetCurrentView("guilds"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("helpBar", int(float32(maxX)*0.05), int(float32(maxY)*0.85), int(float32(maxX)*0.95), int(float32(maxY)*0.90)); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "CTRL+C Quit the GUI | CTRL+D Confirm choices | Space Enable or disable guild/channel")
	}
	return nil
}

func drawGuilds(v *gocui.View) error {
	v.Clear()

	for i := range DiscordGuilds {
		enabledCount := 0

		for _, channel := range DiscordGuilds[i].Channels {
			if channel.Enabled == true {
				enabledCount++
			}
		}
		if enabledCount == len(DiscordGuilds[i].Channels) {
			fmt.Fprintln(v, "[X] "+DiscordGuilds[i].Name)
		} else if enabledCount > 0 && enabledCount < len(DiscordGuilds[i].Channels) {
			fmt.Fprintln(v, "[*] "+DiscordGuilds[i].Name)
		} else {
			fmt.Fprintln(v, "[ ] "+DiscordGuilds[i].Name)
		}

	}
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func cursorDownGuilds(_ *gocui.Gui, v *gocui.View) error {
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

func cursorDownChannels(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		guildInfo := &DiscordGuilds[GuildSelected]
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if ((oy + cy) + 1) < len(guildInfo.Channels) {
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

	g.Update(func(g *gocui.Gui) error {
		guildsView, err := g.View("guilds")
		if err != nil {
			return err
		}
		drawGuilds(guildsView)
		return nil
	})

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
		guildsView, err := g.View("guilds")
		if err != nil {
			return err
		}
		guildsView.SetOrigin(0, GuildSelectedOPos)
		guildsView.SetCursor(0, GuildSelectedCPos)

		selectGuild(g, guildsView)
		channelsView, err := g.View("channels")
		if err != nil {
			return err
		}
		channelsView.SetOrigin(0, oy)
		channelsView.SetCursor(0, cy)

		return nil
	})
	return nil
}

func changeGuildEnabled(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()

	guildInfo := &DiscordGuilds[oy+cy]

	enabledCount := 0

	for i, _ := range guildInfo.Channels {
		channelInfo := &guildInfo.Channels[i]
		if channelInfo.Enabled == true {
			enabledCount++
		}
	}

	if enabledCount == len(guildInfo.Channels) {
		for i, _ := range guildInfo.Channels {
			channelInfo := &guildInfo.Channels[i]
			channelInfo.Enabled = false
		}
	} else {
		for i, _ := range guildInfo.Channels {
			channelInfo := &guildInfo.Channels[i]
			channelInfo.Enabled = true
		}
	}

	g.Update(func(g *gocui.Gui) error {
		guildsView, err := g.View("guilds")
		if err != nil {
			return err
		}
		drawGuilds(guildsView)
		return nil
	})
	return nil
}

func saveName(g *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("modelName", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Enter name for model"
		v.Editable = true

		if _, err := g.SetCurrentView("modelName"); err != nil {
			return err
		}
	}
	return nil
}

func closeConfirm(g *gocui.Gui, _ *gocui.View) error {
	if err := g.DeleteView("modelName"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("guilds"); err != nil {
		return err
	}
	return nil
}

func confirmName(_ *gocui.Gui, v *gocui.View) error {
	var nameContent string
	var err error
	_, cy := v.Cursor()

	if nameContent, err = v.Line(cy); err != nil {
		nameContent = "model"
	}

	ModelName = nameContent

	return gocui.ErrQuit
}
