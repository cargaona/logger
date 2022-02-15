package logger

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/google/uuid"
)

// Logger implements Logger interface.
type Logger struct {
	instanceName  string
	Client        cloudwatchlogsiface.CloudWatchLogsAPI
	LogGroupName  string
	logStreamName string
	DebugLevel    bool
	StdOut        bool
	sequenceToken string
}

// LogDebug sends logs to CloudWatch with the '[DEBUG]' prefix.
func (l *Logger) Debug(format string, input ...interface{}) {
	if l.DebugLevel {
		l.logToCloudWatch(fmt.Sprintf("[DEBUG] %v", input))
	}
}

// LogInfo sends logs to CloudWatch with the '[INFO]' prefix.
func (l *Logger) Info(format string, input ...interface{}) {
	l.logToCloudWatch(fmt.Sprintf("[INFO] %v", input))
}

// logToCloudWatch sends logs to CloudWatch. If required also sends them to stdout.
func (l *Logger) logToCloudWatch(input string) {
	if l.sequenceToken == "" {
		err := l.createNewLogStream()
		if err != nil {
			log.Println("[DEBUG] Error creating a new CloudWatch Stream")
			l.printToStdOut(input)
			return
		}
	}

	messageWithInstance := fmt.Sprintf("[%s] %s", l.instanceName, input)
	message := l.createInputLogEvent(messageWithInstance)

	resp, err := l.Client.PutLogEvents(message)
	if err != nil {
		log.Println("[INFO] Error pushing logs to CloudWatch")
		l.printToStdOut(input)
		return
	}
	if err == nil {
		l.sequenceToken = aws.StringValue(resp.NextSequenceToken)
	}

	l.printToStdOut(input)
}

func (l *Logger) printToStdOut(input interface{}) {
	if l.StdOut {
		log.Println(input)
	}
}

func (l *Logger) createNewLogStream() error {
	name := uuid.New().String()

	_, err := l.Client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(l.LogGroupName),
		LogStreamName: aws.String(name),
	})

	if err != nil {
		return err
	}

	l.logStreamName = name

	return nil
}

func (l *Logger) createInputLogEvent(input interface{}) *cloudwatchlogs.PutLogEventsInput {
	text := fmt.Sprint(input)
	message := &cloudwatchlogs.PutLogEventsInput{
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			{
				Timestamp: aws.Int64(time.Now().UnixNano() / int64(time.Millisecond)),
				Message:   aws.String(text),
			},
		},
		LogGroupName:  aws.String(l.LogGroupName),
		LogStreamName: aws.String(l.logStreamName),
	}

	if l.sequenceToken != "" {
		return message.SetSequenceToken(l.sequenceToken)
	}
	return message
}
