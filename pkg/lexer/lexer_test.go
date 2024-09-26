package lexer

import (
	"encoding/json"
	"fmt"
	"testing"

	// . "github.com/sjhitchner/toolbox/pkg/testing"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type LexerSuite struct{}

var _ = Suite(&LexerSuite{})

func (s *LexerSuite) Test(c *C) {
	lexer := New(MilestoneStr, MilestoneFindFunc)

	for m := range ParseMilestone(lexer) {
		fmt.Println(m)
	}
}

type Milestone struct {
	Name         string
	Goal         string
	Quarter      string
	StartMonth   string
	Description  string
	Risk         string
	Effort       string
	Team         string
	Dependencies string
	Image        string
}

func (t Milestone) String() string {
	b, err := json.Marshal(t)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

const MilestoneStr = `
XXXXXXXXXXXXXXXX\milestone{3.1}{2025Q1}{
	Launch (move) Bid Pacing pacing to GCP.  Pacing will be completely handled in GCP
}{
\item Launch Line Items Aerospike clusters in GCP
	\begin{enumerate}
		\item Setup XDR to replicate namespaces between AWS and GCP
		\begin{enumerate}
			\item line\_items
			\item campaign\_data
			\item bid\_data
		\end{enumerate}
		\item Do not initially replicate bid\_tables between AWS and GCP
	\end{enumerate}
	\item Launch Line Items Publishing
	\begin{enumerate}
		\item K8 utils cluster
		\item PubSub queue per namespace
	\end{enumerate}
	\item Launch pacing system
	\begin{enumerate}
		\item K8 utils cluster
		\item Delivery profiles should be also published to GCS location
		\item Setup forwarding from AWS impressions stream to PubSub impressions stream (branch)
		\item Publish bid\_tables to non-replicated namespace
	\end{enumerate}
	\item Final switch over
	\begin{enumerate}
		\item Turn off AWS pacing
		\item Replicate bid\_tables to AWS
    \end{enumerate}
	\item Shutdown AWS Bid Pacing
}{5}{3}{GCP}{Cloud, DSP, Data Engineering, Data Science}{}


\milestone{3.1}{2025Q1}{
	Launch (move) Bid Pacing pacing to GCP.  Pacing will be completely handled in GCP
}{
\item Launch Line Items Aerospike clusters in GCP
	\begin{enumerate}
		\item Setup XDR to replicate namespaces between AWS and GCP
		\begin{enumerate}
			\item line\_items
			\item campaign\_data
			\item bid\_data
		\end{enumerate}
		\item Do not initially replicate bid\_tables between AWS and GCP
	\end{enumerate}
	\item Launch Line Items Publishing
	\begin{enumerate}
		\item K8 utils cluster
		\item PubSub queue per namespace
	\end{enumerate}
	\item Launch pacing system
	\begin{enumerate}
		\item K8 utils cluster
		\item Delivery profiles should be also published to GCS location
		\item Setup forwarding from AWS impressions stream to PubSub impressions stream (branch)
		\item Publish bid\_tables to non-replicated namespace
	\end{enumerate}
	\item Final switch over
	\begin{enumerate}
		\item Turn off AWS pacing
		\item Replicate bid\_tables to AWS
    \end{enumerate}
	\item Shutdown AWS Bid Pacing
}{5}{3}{GCP}{Cloud, DSP, Data Engineering, Data Science}{}

 `

const (
	StartMilestone = "\\milestone"
	LeftBrace      = '{'
	RightBrace     = '}'

	MilestoneStart TokenType = iota
	MilestoneName
	MilestoneQuarter
	MilestoneStartMonth
	MilestoneGoal
	MilestoneDescription
	MilestoneRisk
	MilestoneEffort
	MilestoneTeam
	MilestoneDependencies
	MilestoneImage
	MilestoneEnd
)

//

func MilestoneFindFunc(l *Lexer) StateFunc {
	for l.pos < len(l.input) {
		if l.Matches(StartMilestone) {
			l.Emit(MilestoneStart)
			return MilestoneNameFunc
		}
		l.Skip()
	}

	fmt.Println("EOF")
	l.Emit(TokenEOF)
	return nil
}

func milestoneInternal(l *Lexer, token TokenType, stateFn StateFunc) StateFunc {
	if l.Next() != LeftBrace {
		l.Emit(TokenError)
		return nil
	}
	l.Ignore()
	l.UntilRune(RightBrace)
	l.Emit(token)
	l.Skip()
	return stateFn
}

func MilestoneNameFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneName, MilestoneQuarterFunc)
}

func MilestoneQuarterFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneQuarter, MilestoneStartMonthFunc)
}

func MilestoneStartMonthFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneStartMonth, MilestoneGoalFunc)
}

func MilestoneGoalFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneGoal, MilestoneDescriptionFunc)
}

func MilestoneDescriptionFunc(l *Lexer) StateFunc {
	fmt.Printf("start(%d) pos(%d) con(%s)\n", l.start, l.pos, l.input[l.start:l.pos])
	if l.Next() != LeftBrace {
		l.Emit(TokenError)
		return nil
	}
	l.Ignore()

	fmt.Printf("START(%d) pos(%d) con(%s) %s\n", l.start, l.pos, l.input[l.start:l.pos], l.input[l.start:])
	depth := 0
	for l.pos < len(l.input) {
		switch l.Next() {
		case LeftBrace:
			fmt.Println("LB", l.input[l.start:l.pos])
			depth++
		case RightBrace:
			fmt.Println("RB", l.input[l.start:l.pos])
			if depth == 0 {
				l.Backup()
				l.Emit(MilestoneDescription)
				l.Skip()
				return MilestoneRiskFunc
			}
			depth--
		}
	}
	fmt.Printf("start(%d) pos(%d) con(%s) %s\n", l.start, l.pos, l.input[l.start:l.pos], l.input[l.start:])

	l.Emit(TokenError)
	return nil
}

func MilestoneRiskFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneRisk, MilestoneEffortFunc)
}

func MilestoneEffortFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneEffort, MilestoneTeamFunc)
}

func MilestoneTeamFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneTeam, MilestoneDependenciesFunc)
}

func MilestoneDependenciesFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneDependencies, MilestoneImageFunc)
}

func MilestoneImageFunc(l *Lexer) StateFunc {
	return milestoneInternal(l, MilestoneImage, MilestoneEndFunc)
}

func MilestoneEndFunc(l *Lexer) StateFunc {
	l.Emit(MilestoneEnd)
	return MilestoneFindFunc
}

/*
	return milestoneInternal(MilestoneDescription, MilestoneGoalFunc)
	fmt.Printf("start(%d) pos(%d) con(%s)\n", l.start, l.pos, l.input[l.start:l.pos])
	fmt.Printf("start(%d) pos(%d) con(%s)\n", l.start, l.pos, l.input[l.start:l.pos])
	if !l.Accept(LeftBrace) {
		l.Emit(TokenError) // Useful to make EOF a token.
		return nil         // Stop the run loop.
	}

	l.Until(RightBrace)
	if l.pos > l.start {
		l.Emit(MilestoneName)
		l.Skip()
		return MilestoneQuarterFunc
	}

	l.Emit(TokenError) // Useful to make EOF a token.
	return nil         // Stop the run loop.
}

	if !l.Accept(LeftBrace) {
		l.Emit(TokenError) // Useful to make EOF a token.
		return nil         // Stop the run loop.
	}

	l.Until(RightBrace)
	if l.pos > l.start {
		l.Emit(MilestoneName)
		l.Skip()
		return MilestoneQuarterFunc
	}

	l.Emit(TokenError) // Useful to make EOF a token.
	return nil         // Stop the run loop.
}

func MilestoneRiskFunc(l *Lexer) StateFunc {
	// GatherName
	return nil
}

func MilestoneEffortFunc(l *Lexer) StateFunc {
	// GatherName
	return nil
}

func MilestoneTeamFunc(l *Lexer) StateFunc {
	// GatherName
	return nil
}

func MilestoneDependencyFunc(l *Lexer) StateFunc {
	// GatherName
	return nil
}

func MilestoneImageFunc(l *Lexer) StateFunc {
	// GatherName
	return nil
}
*/
