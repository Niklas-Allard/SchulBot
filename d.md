require ["fileinto", "mailbox",  "copy"];

if header :contains "Subject" ["#ki", "/ki", "#tasks", "/tasks", "#calendar", "/calendar", "#model", "/model", "#sudoku", "/sudoku", "#news", "/news",  "#hilfe",  "/hilfe", "#help", "/help",  "#translate",  "/translate",  "#übersetze",  "/übersetze"] {
    fileinto :create "INBOX/SchulBot";
}

if anyof (
      header :contains "subject" "#ki",
      header :contains "subject" "/ki",
      header :contains "subject" "#tasks",
      header :contains "subject" "/tasks",
      header :contains "subject" "#calendar",
      header :contains "subject" "/calendar",
      header :contains "subject" "#model",
      header :contains "subject" "/model",
      header :contains "subject" "#sudoku",
      header :contains "subject" "/sudoku",
      header :contains "subject" "#news",
      header :contains "subject" "/news",
      header :contains "subject" "#hilfe",
      header :contains "subject" "/hilfe",
      header :contains "subject" "#help",
      header :contains "subject" "/help",
      header :contains "subject" "#translate",
      header :contains "subject" "/translate",
      header :contains "subject" "#übersetze",
      header :contains "subject" "/übersetze"
  ) {
      redirect :copy "niklas.allard.de@gmail.com";
  }