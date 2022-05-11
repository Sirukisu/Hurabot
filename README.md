# Hurabot

A Discord bot that uses Markov chains to generate random text from your messages.

## 🔧 List of features
- Make word models from Discord messages data
- Generate random text from these models
- Launch a Discord bot that can generate messages with a slash command
- Restrict Discord bot commands to a specific guild only
- Set a maximum amount of words that can be generated with a single command with the Discord bot


## 📦 Installation & Usage

Either:

- Download the [latest release](https://github.com/Sirukisu/Hurabot/releases/latest) & place it to a folder

Or build from source:

1. Install [Go](https://go.dev/) for your system
2. Clone the repository
3. Run `go build` in the source directory


### Configuration
Create a new config using the `bot config create` command.

The current config options are as follows:

| Option              | Description                                                                |
|---------------------|----------------------------------------------------------------------------|
| AuthenticationToken | Bot token for logging into Discord.                                        |
| GuildID             | ID of the guild to register commands to, registers globally if left empty. |
| ModelFolder         | Folder that contains the word models to use.                               |
| ModelsToUse         | List of model files to use if the whole model directory isn't wanted.      |
| MaxWords            | Max amount of words that the bot can generate.                             |
| LogDir              | Directory where to save log files.                                         |
| LogLevel            | Level of logging.                                                          |

### Making word models

1. To make a word model from your messages you first need to [request your data](https://support.discord.com/hc/en-us/articles/360004027692) from Discord
2. Extract the `messages` folder from the .zip file you receive after a few days
3. Create the model with the command `model create -f "</path/to/messages/folder>"`
4. Select what channels you want to include
5. When done, press  CTRL+S and enter a name for the model. This will be displayed in the command choices in Discord.
6. After that, finally enter a filename for the model
7. The model will be saved to the `models` directory at the program's root path


### Creating the Discord bot
1. Create a new application at the [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a [bot user](https://discord.com/developers/docs/topics/oauth2#bots) for the application
3. Generate a token for the bot & add this to the config file
4. Use a OAuth2 link with the bots' application ID, scope of `applications.commands bot` and permissions of `2048` to add the bot to the guild(s) you want


### Running the bot
Run the bot by using the command `bot run`.

Generate text using the `/generate-text` slash command.

## ✍ Features planned

- CUI for managing bot
- Adding timed events for sending daily messages for example
- Make a model from individual messages.csv files
- More bot commands? Ideas are welcome

## ❗ Known issues
- Currently, the `LogLevel` and `ModelsToUse` config options do nothing
- The CUIs can crash the program if the terminal display is too small
- Backslashes can't be entered in the CUIs

## ✨ Credits
- Thanks to my Discord friends who gave me this dumb idea
- Also, this is also my first project that I have managed to do from start to finish, so any feedback on what I could do better is very much welcome!