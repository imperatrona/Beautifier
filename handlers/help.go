package handlers

import "gopkg.in/telebot.v3"

func HelpCommandHandler(c telebot.Context) error {
	return c.Send(
		"‼️ Bot is totaly free and given as is. No support is provided.\n\n"+
			"*How to use*\nGo to twitter, copy link to tweet with video, send it to bot, get video in your messages.\n\nTo use in group you first need to add bot to group, then send command `/tweet@xvideosdwbot TWEET_URL` replace `TWEET_URL` with tweet you want to download.\n\n"+
			"*Settings*\nYou can choose how you want to download tweets, send /settings to read more.\n\n"+
			"*«Video unavailable»*\nMake sure twitter profile is public and tweet is contain video.\n\n"+
			"*Video without audio or doesn't work*\nTwitter changed the way they process videos, you can fix video by reencoding it. You can use [this service](https://ffmpeg-online.vercel.app/?outputOptions=-c%20copy%20-metadata%20m%3Dh) for it",
		&telebot.SendOptions{
			ParseMode:             telebot.ModeMarkdown,
			DisableWebPagePreview: true,
		})
}
