package main

const (
	TEMPLATE_DIR    = DATA_DIR + "templates/" // templates that will be rendered with Go
	STATIC_DIR      = DATA_DIR + "static/"    // static files
	DBFILE          = DATA_DIR + "data.db"
	DATA_DIR        = "/Volume/data/" // location of templates and static files
	PROGRAM_VERSION = "instawidget 0.1a"                                                 // program version

	REDIRECT_URI = "http://sudosu.me:9999/iw/"

	INSTAGRAM_API           = "https://api.instagram.com"
	INSTAGRAM_CLIENT_ID     = "******************************"
	INSTAGRAM_CLIENT_SECRET = "******************************"
	INSTAGRAM_MEDIA_RECENT  = INSTAGRAM_API + "/v1/users/self/media/recent?access_token=%s"                                                                  // API URL for pulling our own data
	INSTAGRAM_OAUTH         = INSTAGRAM_API + "/oauth/authorize/?client_id=" + INSTAGRAM_CLIENT_ID + "&redirect_uri=" + REDIRECT_URI + "&response_type=code" // API URL for OAUTH
)
