package main

import (
	// "fmt"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
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
		instances, err := DescribeInstancesCmd(optionsMapStr, optionsMapStr["instance_id"])
		if err != nil {
			sendMessage(s, i, fmt.Sprintf("Something went wrong...\n```%s```", err))
		} else {
			sendInstanceStatus(s, i, instances, optionsMap)
		}
	},
	"start": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionsMap := getOptionsMapWithCreds(i)
		optionsMapStr := convertMapValuesToString(optionsMap)
		deferMessage(s, i, "test")
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
		deferMessage(s, i, "test")
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
	optionsMap := getOptionsMap(i)
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

func sendInstanceStatus(s *discordgo.Session, i *discordgo.InteractionCreate, instances []map[string]string, options map[string]interface{}) {

	fieldMap := make(map[string][]string)
	for _, instance := range instances {
		for k, v := range instance {
			fieldMap[k] = append(fieldMap[k], v)
		}
	}

	var fields []*discordgo.MessageEmbedField
	for k, v := range fieldMap {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   k,
			Value:  strings.Join(v, "\n"),
			Inline: true,
		})
	}
	sort.Slice(fields, func(a, b int) bool {
		return fields[a].Name < fields[b].Name
	})

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			// Content: content,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Instances Status",
					Description: fmt.Sprintf("Region: `%s`", options["region"]),
					Fields:      fields,
				},
			},
		},
	})
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

func deferMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func deferMessageUpdate(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: content,
	})
}
