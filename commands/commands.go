package commands

import (
	"BetterScorch/execution"
	"BetterScorch/secrets"
	"BetterScorch/sender"
	"fmt"
	"log/slog"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type command struct {
	declaration *discordgo.ApplicationCommand
	handler     func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func AddAllCommands(s *discordgo.Session) {
	fmt.Print("    |   Adding slash commands... ")

	commandSlice := []*discordgo.ApplicationCommand{}
	for _, command := range commands {
		commandSlice = append(commandSlice, command.declaration)
	}
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			defer func() {
				if r := recover(); r != nil {
					sender.Followup(s, i, fmt.Sprintf("PANIC: %v", r))
					sender.Respond(s, i, fmt.Sprintf("PANIC: %v", r), nil)
					slog.Error(fmt.Sprintf("%v", r), "commandName", i.ApplicationCommandData().Name, "commandType", "ApplicationCommand")
					fmt.Println(string(debug.Stack()))
				}
			}()

			if execution.IsDead(i.Member.User.ID) && !isHC(i.Member) {
				sender.RespondEphemeral(s, i, "https://tenor.com/view/yellow-emoji-no-no-emotiguy-no-no-no-gif-gif-9742000569423889376", nil)
			} else {
				slog.Info("Received Command", "name", i.ApplicationCommandData().Name, "commandType", "ApplicationCommand", "args", i.ApplicationCommandData().Options)
				if h, ok := commands[i.ApplicationCommandData().Name]; ok {
					h.handler(s, i)
				}
			}
		}
	})
	_, err := s.ApplicationCommandBulkOverwrite(s.State.Application.ID, secrets.GuildID, commandSlice)
	if err != nil {
		slog.Error(err.Error())
	}
	fmt.Println("Done")

	fmt.Print("    |   Adding component commands... ")
	for commandName, commandFunction := range componentAndModalCommands {
		s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Type == discordgo.InteractionMessageComponent && strings.HasPrefix(strings.ToLower(i.MessageComponentData().CustomID), strings.ToLower(commandName)) {
				defer func() {
					if r := recover(); r != nil {
						sender.Followup(s, i, fmt.Sprintf("PANIC: %v", r))
						sender.Respond(s, i, fmt.Sprintf("PANIC: %v", r), nil)
						slog.Error(fmt.Sprintf("%v", r), "commandName", i.MessageComponentData().CustomID, "commandType", "Component")
						fmt.Println(string(debug.Stack()))
					}
				}()

				slog.Info("Received Command", "name", i.MessageComponentData().CustomID, "commandType", "Component")
				if execution.IsDead(i.Member.User.ID) && !isHC(i.Member) {
					sender.RespondEphemeral(s, i, "https://tenor.com/view/yellow-emoji-no-no-emotiguy-no-no-no-gif-gif-9742000569423889376", nil)
				} else {
					commandFunction(s, i)
				}
			} else if i.Type == discordgo.InteractionModalSubmit && strings.HasPrefix(strings.ToLower(i.ModalSubmitData().CustomID), strings.ToLower(commandName)) {
				defer func() {
					if r := recover(); r != nil {
						sender.Followup(s, i, fmt.Sprintf("PANIC: %v", r))
						sender.Respond(s, i, fmt.Sprintf("PANIC: %v", r), nil)
						slog.Error(fmt.Sprintf("%v", r), "commandName", i.ModalSubmitData().CustomID, "commandType", "Modal")
						fmt.Println(string(debug.Stack()))
					}
				}()

				slog.Info("Received Command: ", "name", i.ModalSubmitData().CustomID, "commandType", "Modal")
				if execution.IsDead(i.Member.User.ID) && !isHC(i.Member) {
					sender.RespondEphemeral(s, i, "https://tenor.com/view/yellow-emoji-no-no-emotiguy-no-no-no-gif-gif-9742000569423889376", nil)
				} else {
					commandFunction(s, i)
				}
			}
		})
	}
	fmt.Println("Done")

	fmt.Print("    |   Adding autocompletions... ")
	for commandName, commandFunction := range autocompletes {
		s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Type == discordgo.InteractionApplicationCommandAutocomplete && strings.HasPrefix(strings.ToLower(i.ApplicationCommandData().Name), strings.ToLower(commandName)) {
				defer func() {
					if r := recover(); r != nil {
						sender.Followup(s, i, fmt.Sprintf("PANIC: %v", r))
						sender.Respond(s, i, fmt.Sprintf("PANIC: %v", r), nil)
						slog.Error(fmt.Sprintf("%v", r), "commandName", i.ApplicationCommandData().Name, "commandType", "Autocomplete")
						fmt.Println(string(debug.Stack()))
					}
				}()

				slog.Info("Received Command: ", "name", i.ApplicationCommandData().Name, "commandType", "Autocomplete")
				commandFunction(s, i)
			}
		})
	}
	fmt.Println("Done")
}

func isHC(m *discordgo.Member) bool {
	if IsAdminAbuser(m) {
		return true
	}

	var roles = []string{"1195135956471255140", "1195858311627669524", "1195858271349784639", "1195136106811887718"}

	for _, role := range m.Roles {
		if slices.Contains(roles, role) {
			return true
		}
	}
	return false
}

func isSWAG(m *discordgo.Member) bool {
	if slices.Contains(m.Roles, "1199148174258995231") {
		return true
	}
	return false
}

func IsAdminAbuser(m *discordgo.Member) bool {
	return m.User.ID == "384422339393355786" || m.User.ID == "394212669114286090" || m.User.ID == "920342100468436993" || m.User.ID == "1079774043684745267" || m.User.ID == "952145898824138792" || m.User.ID == "455833801638281216"
}
