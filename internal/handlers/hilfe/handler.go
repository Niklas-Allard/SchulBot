package hilfe

import (
	"context"
	"strings"

	"schulbot/internal/model"
)

type Handler struct{}

func New() *Handler { return &Handler{} }

func (h *Handler) CanHandle(tag string) bool {
	return tag == "#hilfe" || tag == "/hilfe" ||
		tag == "#help" || tag == "/help"
}

func (h *Handler) Handle(_ context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	return model.CommandResponse{
		ReplySubject: reply(req.Subject),
		ReplyBody:    helpText,
	}, nil
}

func reply(subject string) string {
	if strings.HasPrefix(subject, "Re: ") {
		return subject
	}
	return "Re: " + subject
}

const helpText = `SchulBot - Befehlsuebersicht

*** KURZUEBERSICHT ***

#ki <Frage>              KI-Antwort (Cloudflare / Gemini)
/model [Nr.|Name|reset]  KI-Modell wechseln
/news [Thema|suche]      Tagesschau-Nachrichten
/sudoku [Schwierigkeit]  Sudoku-Raetsel als Bild
/tasks [liste|add|...]   Google Tasks
/calendar [heute|neu]    Google Kalender
/hilfe                   Diese Hilfe


*** DETAILS ***

-- #ki --
KI beantwortet deine Frage oder Aufgabe.

  #ki Was ist die Relativitaetstheorie?
  /ki Schreibe ein Gedicht ueber den Herbst

Ohne Text nach dem Tag: Kurzanleitung wird gesendet.
Modell aendern mit /model.


-- /model --
Waehlt das KI-Modell fuer deine #ki-Anfragen.
Die Wahl gilt nur fuer dich und bleibt gespeichert.

  /model         -> Liste aller Modelle
  /model 2       -> Modell Nr. 2 waehlen
  /model reset   -> Zurueck auf Standard


-- /news --
Aktuelle Nachrichten von tagesschau.de (kein Account noetig).

  /news                -> Top-Nachrichten
  /news inland         -> Deutschland
  /news ausland        -> International
  /news sport          -> Sport
  /news wirtschaft     -> Wirtschaft
  /news wissenschaft   -> Wissenschaft & Tech
  /news gesundheit     -> Gesundheit
  /news faktenfinder   -> Faktenfinder
  /news region nrw     -> Regional (Name oder Nr. 1-16)
  /news suche Klima    -> Volltextsuche

Regionen: 1=BW 2=BY 3=BE 4=BB 5=HB 6=HH 7=HE
          8=MV 9=NI 10=NW 11=RP 12=SL 13=SN
          14=ST 15=SH 16=TH


-- /sudoku --
Sudoku-Raetsel als PNG-Bild im Anhang.
Die ID ist 48 Stunden gueltig.

  /sudoku              -> Neues Raetsel (Einfach)
  /sudoku medium       -> Mittel
  /sudoku hard         -> Schwer
  /sudoku loesung ID   -> Loesungsbild abrufen


-- /tasks --
Google Tasks verwalten.
Benoetigt: GOOGLE_CLIENT_ID, CLIENT_SECRET, REFRESH_TOKEN.

  /tasks liste          -> Offene Aufgaben
  /tasks add Titel      -> Neue Aufgabe
  /tasks fertig Titel   -> Als erledigt markieren


-- /calendar --
Google Kalender (gleiche Voraussetzungen wie /tasks).

  /calendar heute                  -> Heutige Termine
  /calendar woche                  -> Diese Woche
  /calendar neu Sport morgen 14:00 -> Termin anlegen

Datum: heute / morgen / JJJJ-MM-TT / TT.MM.JJJJ
Uhrzeit optional, Standard: 08:00.


-- /hilfe --
Diese Uebersicht. Auch per #help oder /help erreichbar.


Tags koennen im Betreff ODER in der ersten Zeile des
E-Mail-Textes stehen. Gross-/Kleinschreibung egal.`
