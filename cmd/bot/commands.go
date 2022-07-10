package main

import (
	// "fmt"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/ecoshub/stable"
)

var regionList = [...]string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ca-central-1"}

// Commands, Options, Choices
var commands = []*discordgo.ApplicationCommand{

	{
		Name:        "help",
		Description: "ValBot Help",
	},
	{
		Name:        "init",
		Description: "Initialize ValBot",
		Options: []*discordgo.ApplicationCommandOption{
			regionOption,
			{
				Name:        "aws_access_key_id",
				Description: "AWS Access Key ID",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "aws_secret_access_key",
				Description: "AWS Secret Access Key",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	},
	{
		Name:        "init-delete",
		Description: "Delete Initialized Credential from ValBot",
		Options: []*discordgo.ApplicationCommandOption{
			regionOption,
		},
	},
	{
		Name:        "status",
		Description: "Servers Status",
		Options: []*discordgo.ApplicationCommandOption{
			regionOption,
			// {
			// 	Name:        "instance_id",
			// 	Description: "Instance ID",
			// 	Type:        discordgo.ApplicationCommandOptionString,
			// 	Required:    false,
			// },
		},
	},
	{
		Name:        "start",
		Description: "Start Servers",
		Options: []*discordgo.ApplicationCommandOption{
			regionOption,
			{
				Name:        "instance_id",
				Description: "Instance ID",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	},
	{
		Name:        "stop",
		Description: "Stop Servers",
		Options: []*discordgo.ApplicationCommandOption{
			regionOption,
			{
				Name:        "instance_id",
				Description: "Instance ID",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	},
}

var regionOption = &discordgo.ApplicationCommandOption{
	Name:        "region",
	Description: "AWS Region",
	Type:        discordgo.ApplicationCommandOptionString,
	Required:    true,
	Choices:     getRegionChoices(),
}

func getRegionChoices() []*discordgo.ApplicationCommandOptionChoice {
	var choices []*discordgo.ApplicationCommandOptionChoice

	for _, r := range regionList {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  r,
			Value: r,
		})
	}
	return choices
}

// Command Handlers
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		sendMessage(s, i, "Setup Valbot with `/init` to use the other commands. AWS Region must be specified for all commands.")
	},
	"init": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionsMap := getOptionsMap(i)

		err := saveCredsToDB(optionsMap)
		if err != nil {
			log.Println(err)
			sendMessageEphemeral(s, i, fmt.Sprintf("Something went wrong...\n```%s```", err))
		} else {
			sendMessageEphemeral(s, i, fmt.Sprintf("Initialized ValBot for Guild ID: `%s` Region: `%s`", optionsMap["guild_id"], optionsMap["region"]))
		}
	},
	"init-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionsMap := getOptionsMap(i)
		// data := getCredsFromDB(optionsMap)
		deleteDB(optionsMap)
		sendMessageEphemeral(s, i, fmt.Sprintf("Deleted ValBot AWS Credentials for Guild ID: `%s` Region: `%s`", optionsMap["guild_id"], optionsMap["region"]))
	},
	"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionsMap := getOptionsMapWithCreds(i)
		optionsMapStr := convertMapValuesToString(optionsMap)
		deferMessageStatus(s, i)
		instances, err := DescribeInstancesCmd(optionsMapStr, optionsMapStr["instance_id"])
		if err != nil {
			deferMessageUpdate(s, i, fmt.Sprintf("Something went wrong...\n```%s```", err))
		} else {
			sendInstanceStatus(s, i, instances, optionsMap)
		}
	},
	"start": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionsMap := getOptionsMapWithCreds(i)
		optionsMapStr := convertMapValuesToString(optionsMap)
		deferMessage(s, i)
		err := StartInstancesCmd(optionsMapStr, optionsMapStr["instance_id"])
		if err != nil {
			deferMessageUpdate(s, i, fmt.Sprintf("Something went wrong...\n```%s```", err))
		} else {

			deferMessageUpdate(s, i, fmt.Sprintf("Starting instance `%s` in `%s`. Check `/status region: %s` to see more info.", optionsMapStr["instance_id"], optionsMapStr["region"], optionsMapStr["region"]))
		}
	},
	"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionsMap := getOptionsMapWithCreds(i)
		optionsMapStr := convertMapValuesToString(optionsMap)
		deferMessage(s, i)
		err := StopInstancesCmd(optionsMapStr, optionsMapStr["instance_id"])
		if err != nil {
			deferMessageUpdate(s, i, fmt.Sprintf("Something went wrong...\n```%s```", err))
		} else {
			deferMessageUpdate(s, i, fmt.Sprintf("Stopping instance `%s` in `%s`. Check `/status region: %s` to see more info.", optionsMapStr["instance_id"], optionsMapStr["region"], optionsMapStr["region"]))
		}
	},
}

// Helper functions
func getOptionsMap(i *discordgo.InteractionCreate) map[string]interface{} {
	options := i.ApplicationCommandData().Options
	optionsMap := make(map[string]interface{})
	optionsMap["guild_id"] = i.GuildID

	for _, opt := range options {
		optionsMap[opt.Name] = opt.StringValue()
	}
	return optionsMap
}

func getOptionsMapWithCreds(i *discordgo.InteractionCreate) map[string]interface{} {
	options := i.ApplicationCommandData().Options
	optionsMap := make(map[string]interface{})
	optionsMap["guild_id"] = i.GuildID

	for _, opt := range options {
		optionsMap[opt.Name] = opt.StringValue()
	}

	data := getCredsFromDB(optionsMap)
	for _, d := range data {
		for k, v := range d {
			optionsMap[k] = v
		}
	}
	return optionsMap
}

func convertMapValuesToString(input map[string]interface{}) map[string]string {
	output := make(map[string]string)

	for k, v := range input {
		output[k] = v.(string)
	}
	return output
}

func sendMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func sendMessageEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   uint64(discordgo.MessageFlagsEphemeral),
		},
	})
}

func deferMessage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func deferMessageUpdate(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: content,
	})
}

func deferMessageStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Instances Status",
				},
			},
		},
	})
}

func sendInstanceStatus(s *discordgo.Session, i *discordgo.InteractionCreate, instances []map[string]interface{}, options map[string]interface{}) {

	table, err := stable.ToTable(instances)
	if err != nil {
		fmt.Println(err)
		return
	}
	table.SetCaption(fmt.Sprintf("Status - %s", options["region"]))

	deferMessageUpdate(s, i, fmt.Sprintf("```\n%s```", table.String()))
}
