package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
	"os"
	"strconv"
)

var ConfigNoSave bool = false

// ConfigEditCUI open CUI for editing configs
func ConfigEditCUI(configFile *os.File) {
	// check if file was provided
	if configFile == nil {
		if LoadedConfig == nil {
			log.Fatalln("No config was loaded")
		}
		// config already loaded, don't load it again
	} else {
		if err := ConfigLoadConfig(configFile); err != nil {
			log.Fatalf("Failed to load config from %s: %s", configFile.Name(), err.Error())
		}
	}

	// initialize CUI
	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		log.Panicln(err)
	}

	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(configEditLayout)

	// keybinding for cursor down in config options view
	if err := g.SetKeybinding("configOptions", gocui.KeyArrowDown, gocui.ModNone, configEditCursorDown); err != nil {
		log.Panicln(err)
	}
	// keybinding for cursor up
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, configEditCursorUp); err != nil {
		log.Panicln(err)
	}
	// keybinding for handling enter presses
	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, configEditEnterHandler); err != nil {
		log.Panicln(err)
	}
	// keybinding for cursor down in log level edit view
	if err := g.SetKeybinding("editLogLevel", gocui.KeyArrowDown, gocui.ModNone, configEditLogLevelCursorDown); err != nil {
		log.Panicln(err)
	}
	// keybinding for cursor down in models to use view
	if err := g.SetKeybinding("editModelsToUse", gocui.KeyArrowDown, gocui.ModNone, configEditModelsToUseCursorDown); err != nil {
		log.Panicln(err)
	}
	// keybinding for adding new model to use in models to use view
	if err := g.SetKeybinding("editModelsToUse", gocui.KeySpace, gocui.ModNone, configEditModelsToUseAddModel); err != nil {
		log.Panicln(err)
	}
	// keybinding for removing a model in models to use view
	if err := g.SetKeybinding("editModelsToUse", gocui.KeyDelete, gocui.ModNone, configEditModelsToUseRemoveModel); err != nil {
		log.Panicln(err)
	}
	// keybinding for quiting the CUI
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, gocuiQuitWithoutSave); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

// Main layout function for config CUI
func configEditLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("configOptions", int(float32(maxX)*0.05), 0, int(float32(maxX)*0.95), int(float32(maxY)*0.7)); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
		v.Title = "Set config options"

		drawOptions(v)

		if _, err := g.SetCurrentView("configOptions"); err != nil {
			return err
		}

	}
	return nil
}

// Print the config options to a gocui.View
func drawOptions(v *gocui.View) {
	v.Clear()

	fmt.Fprintf(v, "Authentication token: %s\n"+
		"Guild ID: %s\n"+
		"Models directory: %s\n"+
		"Models to use: %d total\n"+
		"Maximum amount of words: %d\n"+
		"Log directory: %s\n"+
		"Log level: %s\n"+
		"Save config",
		LoadedConfig.AuthenticationToken, LoadedConfig.GuildID, LoadedConfig.ModelDirectory, len(LoadedConfig.ModelsToUse),
		LoadedConfig.MaxWords, LoadedConfig.LogDir, LoadedConfig.LogLevel)
}

// Print the models to use to a gocui.View
func drawModelsToUse(v *gocui.View) {
	v.Clear()

	for i := range LoadedConfig.ModelsToUse {
		fmt.Fprintln(v, LoadedConfig.ModelsToUse[i])
	}
}

// Function for cursor down in config edit view
func configEditCursorDown(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if ((oy + cy) + 1) < 5 {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Function for cursor up
func configEditCursorUp(_ *gocui.Gui, v *gocui.View) error {
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

// Function for cursor down in config log level view
func configEditLogLevelCursorDown(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if ((oy + cy) + 1) < 2 {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Function for handling enter presses in config CUI
func configEditEnterHandler(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()

	// check the view
	switch v.Name() {
	// open the editor views for the respective options
	case "configOptions":

		maxX, maxY := g.Size()

		optionSelected := oy + cy

		switch optionSelected {
		// open authentication token edit view
		case 0:
			if v, err := g.SetView("editAuthenticationToken", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				v.Title = "Edit Discord Authentication Token"
				v.Editable = true

				fmt.Fprint(v, LoadedConfig.AuthenticationToken)

				if v, err := g.SetView("helpBar", int(float32(maxX)*0.05), int(float32(maxY)*0.75), int(float32(maxX)*0.95), int(float32(maxY)*0.80)); err != nil {
					if err != gocui.ErrUnknownView {
						return err
					}
					fmt.Fprintln(v, "Bot token for logging into Discord")
				}

				if _, err := g.SetCurrentView("editAuthenticationToken"); err != nil {
					return err
				}
			}
		// open guild ID edit view
		case 1:
			if v, err := g.SetView("editGuildName", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				v.Title = "Edit Guild ID"
				v.Editable = true

				fmt.Fprint(v, LoadedConfig.GuildID)

				if v, err := g.SetView("helpBar", int(float32(maxX)*0.05), int(float32(maxY)*0.85), int(float32(maxX)*0.95), int(float32(maxY)*0.90)); err != nil {
					if err != gocui.ErrUnknownView {
						return err
					}
					fmt.Fprintln(v, "ID of the guild to register commands to, registers globally if left empty")
				}

				if _, err := g.SetCurrentView("editGuildName"); err != nil {
					return err
				}
			}
		// open model directory edit view
		case 2:
			if v, err := g.SetView("editModelDirectory", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				v.Title = "Edit Model Directory"
				v.Editable = true

				fmt.Fprint(v, LoadedConfig.ModelDirectory)

				if v, err := g.SetView("helpBar", int(float32(maxX)*0.05), int(float32(maxY)*0.85), int(float32(maxX)*0.95), int(float32(maxY)*0.90)); err != nil {
					if err != gocui.ErrUnknownView {
						return err
					}
					fmt.Fprintln(v, "Folder that contains the word models to use")
				}

				if _, err := g.SetCurrentView("editModelDirectory"); err != nil {
					return err
				}
			}
		// open models to use edit view
		case 3:
			if v, err := g.SetView("editModelsToUse", int(float32(maxX)*0.2), int(float32(maxY)*0.05), int(float32(maxX)*0.8), int(float32(maxY)*0.7)); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				v.Title = "Edit Models To Use"
				v.Highlight = true
				v.SelBgColor = gocui.ColorGreen
				v.SelFgColor = gocui.ColorBlack

				drawModelsToUse(v)

				if v, err := g.SetView("editModelsToUseHelpBar", int(float32(maxX)*0.05), int(float32(maxY)*0.75), int(float32(maxX)*0.95), int(float32(maxY)*0.85)); err != nil {
					if err != gocui.ErrUnknownView {
						return err
					}
					fmt.Fprintln(v, "List of model files to use if the whole model directory isn't wanted\n[Enter] close view | [Space] add new model | [Delete] remove model")
				}

				if _, err := g.SetCurrentView("editModelsToUse"); err != nil {
					return err
				}
			}
		// open edit max words view
		case 4:
			if v, err := g.SetView("editMaxWords", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				v.Title = "Edit Maximum Amount of Words"
				v.Editable = true

				fmt.Fprint(v, LoadedConfig.MaxWords)

				if v, err := g.SetView("helpBar", int(float32(maxX)*0.05), int(float32(maxY)*0.85), int(float32(maxX)*0.95), int(float32(maxY)*0.90)); err != nil {
					if err != gocui.ErrUnknownView {
						return err
					}
					fmt.Fprintln(v, "Max amount of words that the bot can generate")
				}

				if _, err := g.SetCurrentView("editMaxWords"); err != nil {
					return err
				}
			}
		// open log directory edit view
		case 5:
			if v, err := g.SetView("editLogDirectory", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				v.Title = "Edit Log Directory"
				v.Editable = true

				fmt.Fprint(v, LoadedConfig.LogDir)

				if v, err := g.SetView("helpBar", int(float32(maxX)*0.05), int(float32(maxY)*0.85), int(float32(maxX)*0.95), int(float32(maxY)*0.90)); err != nil {
					if err != gocui.ErrUnknownView {
						return err
					}
					fmt.Fprintln(v, "Directory where to save log files")
				}

				if _, err := g.SetCurrentView("editLogDirectory"); err != nil {
					return err
				}
			}
		// open log level edit view
		case 6:
			if v, err := g.SetView("editLogLevel", maxX/2-30, maxY/2, maxX/2+30, maxY/2+3); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				v.Title = "Select Log Level"
				v.Highlight = true
				v.SelBgColor = gocui.ColorCyan
				v.SelFgColor = gocui.ColorBlack

				fmt.Fprintln(v, "default")
				fmt.Fprintln(v, "verbose")

				if v, err := g.SetView("helpBar", int(float32(maxX)*0.05), int(float32(maxY)*0.85), int(float32(maxX)*0.95), int(float32(maxY)*0.90)); err != nil {
					if err != gocui.ErrUnknownView {
						return err
					}
					fmt.Fprintln(v, "Level of logging")
				}

				if _, err := g.SetCurrentView("editLogLevel"); err != nil {
					return err
				}
			}
		// quit the CUI & save
		case 7:
			return gocui.ErrQuit
		}
		return nil

	// handlers for the editor views
	case "editAuthenticationToken":
		if option, err := v.Line(cy); err != nil {
			v.Clear()
			fmt.Fprint(v, "Invalid option: "+err.Error())
			return nil
		} else {
			LoadedConfig.AuthenticationToken = option
		}

		if err := g.DeleteView("editAuthenticationToken"); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("configOptions"); err != nil {
			return err
		}
		if err := g.DeleteView("helpBar"); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			configOptionsView, err := g.View("configOptions")
			if err != nil {
				return err
			}
			drawOptions(configOptionsView)
			return nil
		})

		return nil

	case "editGuildName":
		if option, err := v.Line(cy); err != nil {
			v.Clear()
			fmt.Fprint(v, "Invalid option: "+err.Error())
			return nil
		} else {
			LoadedConfig.GuildID = option
		}

		if err := g.DeleteView("editGuildName"); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("configOptions"); err != nil {
			return err
		}
		if err := g.DeleteView("helpBar"); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			configOptionsView, err := g.View("configOptions")
			if err != nil {
				return err
			}
			drawOptions(configOptionsView)
			return nil
		})

		return nil

	case "editModelDirectory":
		if option, err := v.Line(cy); err != nil {
			v.Clear()
			fmt.Fprint(v, "Invalid option: "+err.Error())
			return nil
		} else {
			LoadedConfig.ModelDirectory = option
		}

		if err := g.DeleteView("editModelDirectory"); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("configOptions"); err != nil {
			return err
		}
		if err := g.DeleteView("helpBar"); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			configOptionsView, err := g.View("configOptions")
			if err != nil {
				return err
			}
			drawOptions(configOptionsView)
			return nil
		})

		return nil

	case "editModelsToUse":

		if err := g.DeleteView("editModelsToUse"); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("configOptions"); err != nil {
			return err
		}

		if err := g.DeleteView("editModelsToUseHelpBar"); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			configOptionsView, err := g.View("configOptions")
			if err != nil {
				return err
			}
			drawOptions(configOptionsView)
			return nil
		})

		return nil

	case "editMaxWords":
		if option, err := v.Line(cy); err != nil {
			v.Clear()
			fmt.Fprint(v, "Invalid option: "+err.Error())
			return nil
		} else {
			maxWords, err := strconv.Atoi(option)
			if err != nil {
				v.Clear()
				fmt.Fprint(v, "Invalid integer: "+err.Error())
				return nil
			}
			LoadedConfig.MaxWords = maxWords
		}

		if err := g.DeleteView("editMaxWords"); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("configOptions"); err != nil {
			return err
		}
		if err := g.DeleteView("helpBar"); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			configOptionsView, err := g.View("configOptions")
			if err != nil {
				return err
			}
			drawOptions(configOptionsView)
			return nil
		})

		return nil

	case "editModelsToUseAddNew":
		if option, err := v.Line(cy); err != nil {
			v.Clear()
			fmt.Fprint(v, "Invalid option: "+err.Error())
			return nil
		} else {
			LoadedConfig.ModelsToUse = append(LoadedConfig.ModelsToUse, option)
		}

		if err := g.DeleteView("editModelsToUseAddNew"); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("editModelsToUse"); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			editModelsToUseView, err := g.View("editModelsToUse")
			if err != nil {
				return err
			}
			drawModelsToUse(editModelsToUseView)
			return nil
		})

		return nil

	case "editLogDirectory":
		if option, err := v.Line(cy); err != nil {
			v.Clear()
			fmt.Fprint(v, "Invalid option: "+err.Error())
			return nil
		} else {
			LoadedConfig.LogDir = option
		}

		if err := g.DeleteView("editLogDirectory"); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("configOptions"); err != nil {
			return err
		}
		if err := g.DeleteView("helpBar"); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			configOptionsView, err := g.View("configOptions")
			if err != nil {
				return err
			}
			drawOptions(configOptionsView)
			return nil
		})

		return nil

	case "editLogLevel":
		if option, err := v.Line(cy); err != nil {
			v.Clear()
			fmt.Fprint(v, "Invalid option: "+err.Error())
			return nil
		} else {
			LoadedConfig.LogLevel = option
		}

		if err := g.DeleteView("editLogLevel"); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("configOptions"); err != nil {
			return err
		}
		if err := g.DeleteView("helpBar"); err != nil {
			return err
		}

		g.Update(func(g *gocui.Gui) error {
			configOptionsView, err := g.View("configOptions")
			if err != nil {
				return err
			}
			drawOptions(configOptionsView)
			return nil
		})

		return nil

	}
	return nil
}

// Function for cursor down in models to use view
func configEditModelsToUseCursorDown(_ *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if ((oy + cy) + 1) < len(LoadedConfig.ModelsToUse) {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Function for adding a new model to the models to use slice
func configEditModelsToUseAddModel(g *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("editModelsToUseAddNew", maxX/2-30, maxY/2, maxX/2+30, maxY/2+3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Enter path of model"
		v.Editable = true

		if _, err := g.SetCurrentView("editModelsToUseAddNew"); err != nil {
			return err
		}
	}
	return nil
}

// Function for removing a model from the models to use slice
func configEditModelsToUseRemoveModel(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()

	modelSelected := cy + oy

	if _, err := v.Line(cy); err != nil {
		return nil
	}

	if len(LoadedConfig.ModelsToUse) <= modelSelected {
		return nil
	}

	LoadedConfig.ModelsToUse = configEditModelsToUseRemoveFromSlice(LoadedConfig.ModelsToUse, modelSelected)

	g.Update(func(g *gocui.Gui) error {
		editModelsToUseView, err := g.View("editModelsToUse")
		if err != nil {
			return err
		}
		drawModelsToUse(editModelsToUseView)
		return nil
	})

	return nil
}

// Function for removing an index from a slice
func configEditModelsToUseRemoveFromSlice(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func gocuiQuitWithoutSave(_ *gocui.Gui, _ *gocui.View) error {
	ConfigNoSave = true
	return gocui.ErrQuit
}
