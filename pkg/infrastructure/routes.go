package infrastructure

func (s Server) Routes() {
	s.Router.Handle("/", s.handleHomePage())
	s.Router.HandleFunc("/snippet", s.handleDisplaySnippet())
	s.Router.HandleFunc("/snippet/create", s.handleCreateSnippet())
}
