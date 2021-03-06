package discord

import (
	"fmt"
	"strings"
	"sync"

	"github.com/biohuns/discord-servertool/entity"
	"github.com/biohuns/discord-servertool/util"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/xerrors"
)

// Service Discordサービス
type Service struct {
	log       entity.LogService
	instance  entity.InstanceService
	server    entity.ServerStatusService
	session   *discordgo.Session
	channelID string
	botID     string
}

// Start ハンドラを追加して監視を開始
func (s *Service) Start() error {
	s.session.AddHandler(s.newHandler())

	if err := s.session.Open(); err != nil {
		return xerrors.Errorf("failed to open session: %w", err)
	}

	return nil
}

// Send メッセージ送信
func (s *Service) Send(userID, msg string) error {
	if userID != "" {
		msg = fmt.Sprintf("<@!%s>\n%s", userID, msg)
	}

	if _, err := s.session.ChannelMessageSend(s.channelID, msg); err != nil {
		return xerrors.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (s *Service) newHandler() func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(sess *discordgo.Session, m *discordgo.MessageCreate) {
		if !s.isCommand(m) {
			return
		}

		switch s.getCommand(m) {
		// インスタンス起動
		case "start":
			if err := s.instance.Start(); err != nil {
				_ = s.Send(m.Author.ID, fmt.Sprintf("```Failed to Start Instance``````%+v```", err))
				s.log.Error(xerrors.Errorf("failed to start instance: %w", err))
			}
			_ = s.Send(m.Author.ID, "```Starting Instance...```")

		// インスタンス停止
		case "stop":
			if err := s.instance.Stop(); err != nil {
				_ = s.Send(m.Author.ID, fmt.Sprintf("```Failed to Stop Instance``````%+v```", err))
				s.log.Error(xerrors.Errorf("failed to stop instance: %w", err))
			}
			_ = s.Send(m.Author.ID, "```Stopping Instance...```")

		// インスタンスステータス取得
		case "status":
			instanceStatus, err := s.instance.GetCachedStatus()
			if err != nil {
				_ = s.Send(m.Author.ID, fmt.Sprintf("```Failed to Get Instance Status``````%+v```", err))
				s.log.Error(xerrors.Errorf("failed to get instance status: %w", err))
			}

			serverStatus, err := s.server.GetCachedStatus()
			if err != nil {
				_ = s.Send(m.Author.ID, fmt.Sprintf("```Failed to Get Server Status``````%+v```", err))
				s.log.Error(xerrors.Errorf("failed to get server status: %w", err))
			}

			_ = s.Send(m.Author.ID,
				util.InstanceStatusText(
					instanceStatus.Name,
					instanceStatus.StatusCode.String(),
				)+util.ServerStatusText(
					serverStatus.IsOnline,
					serverStatus.GameName,
					serverStatus.PlayerCount,
					serverStatus.MaxPlayerCount,
					serverStatus.Map,
				),
			)

		default:
			_ = s.Send(m.Author.ID, "```start:  Start Instance\nstop:   Stop Instance\nstatus: Get Instance Status```")
		}
	}
}

func (s *Service) getCommand(m *discordgo.MessageCreate) string {
	cmd := strings.TrimSpace(m.Content)

	if strings.HasPrefix(cmd, fmt.Sprintf("<@%s>", s.botID)) {
		cmd = strings.Replace(cmd, fmt.Sprintf("<@%s>", s.botID), "", 1)
	} else if strings.HasPrefix(cmd, fmt.Sprintf("<@!%s>", s.botID)) {
		cmd = strings.Replace(cmd, fmt.Sprintf("<@!%s>", s.botID), "", 1)
	} else {
		return ""
	}

	return strings.TrimSpace(cmd)
}

func (s *Service) isCommand(m *discordgo.MessageCreate) bool {
	return s.botID != m.Author.ID &&
		m.ChannelID == s.channelID &&
		(strings.HasPrefix(m.Content, fmt.Sprintf("<@%s>", s.botID)) ||
			strings.HasPrefix(m.Content, fmt.Sprintf("<@!%s>", s.botID))) &&
		s.getCommand(m) != ""
}

var (
	shared *Service
	once   sync.Once
)

// ProvideService サービス返却
func ProvideService(
	log entity.LogService,
	conf entity.ConfigService,
	instance entity.InstanceService,
	server entity.ServerStatusService,
) (entity.MessageService, error) {
	var err error

	once.Do(func() {
		var session *discordgo.Session
		session, err = discordgo.New()
		if err != nil {
			err = xerrors.Errorf("failed to create session: %w", err)
			return
		}

		session.Token = fmt.Sprintf("Bot %s", conf.Config().Discord.Token)

		shared = &Service{
			log:       log,
			instance:  instance,
			server:    server,
			session:   session,
			channelID: conf.Config().Discord.ChannelID,
			botID:     conf.Config().Discord.BotID,
		}
	})

	if err != nil {
		return nil, xerrors.Errorf("failed to provide service: %w", err)
	}

	if shared == nil {
		return nil, xerrors.New("service is not provided")
	}

	return shared, nil
}
