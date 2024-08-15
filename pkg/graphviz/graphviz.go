package graphviz

import (
	"fmt"
	"sort"
	"strings"
)

const Tab = "  "

type Direction string

const (
	Forward Direction = "forward"
	Both              = "both"
)

type ImageType int

const (
	None ImageType = iota
	Athena
	DynamoDB
	ElasticBeanstalk
	Firehose
	Flink
	Kinesis
	Lambda
	LoadBalancer
	RDS
	S3
	SNS
	SQS
	Question
	BiqQuery
	Cloud
	Coralogix
	Datadog
	Redis
)

var legendCh chan Legend

type Legend struct {
	Label string
	Image string
}

func Prefix(prefix, id string) string {
	if prefix == "" {
		return id
	}
	return prefix + "_" + id
}

func Indent(indent int) string {
	return strings.Repeat(Tab, indent)
}

func Quote(i interface{}) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.3f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		return fmt.Sprintf("\"%s\"", v)
	}
	return ""
}

func Attributes(m map[string]interface{}, attrs ...string) string {
	for k, v := range m {
		attrs = append(attrs, fmt.Sprintf("%s=%s", k, Quote(v)))
	}
	return "[" + strings.Join(attrs, ",") + "]"
}

func Image(id string, image ImageType) string {
	if strings.Contains(id, "aws") {
		awsMap := map[ImageType]Legend{
			Athena:           {Label: "Athena", Image: "icons/aws/athena.png"},
			DynamoDB:         {Label: "DynamoDB", Image: "icons/aws/dynamodb.png"},
			ElasticBeanstalk: {Label: "Elastic Beanstalk", Image: "icons/aws/eb.png"},
			Firehose:         {Label: "Firehose", Image: "icons/aws/firehose.png"},
			Flink:            {Label: "Flink", Image: "icons/aws/flink.png"},
			Kinesis:          {Label: "Kinesis", Image: "icons/aws/kinesis.png"},
			Lambda:           {Label: "Lambda", Image: "icons/aws/lambda.png"},
			LoadBalancer:     {Label: "Load Balancer", Image: "icons/aws/lb.png"},
			RDS:              {Label: "RDS", Image: "icons/aws/rds.png"},
			S3:               {Label: "S3", Image: "icons/aws/s3.png"},
			SNS:              {Label: "SNS", Image: "icons/aws/sns.png"},
			SQS:              {Label: "SQS", Image: "icons/aws/sqs.png"},
			Question:         {Label: "Question", Image: "icons/question.png"},
			BiqQuery:         {Label: "BiqQuery", Image: "icons/aws/bigquery.png"},
			Cloud:            {Label: "Cloud", Image: "icons/cloud.png"},
			Coralogix:        {Label: "Coralogix", Image: "icons/coralogix.svg"},
			Datadog:          {Label: "Datadog", Image: "icons/datadog.png"},
			Redis:            {Label: "Redis", Image: "icons/redis.png"},
		}
		legend := awsMap[image]
		legendCh <- legend
		return legend.Image
	}

	gcpMap := map[ImageType]Legend{
		Athena:       {Label: "Athena", Image: "icons/gcp/bigquery.png"},
		DynamoDB:     {Label: "Big Table", Image: "icons/gcp/bigtable.png"},
		Firehose:     {Label: "Pub Sub", Image: "icons/gcp/pubsub.png"},
		Flink:        {Label: "Flink", Image: "icons/gcp/flink.png"},
		Kinesis:      {Label: "Pub Sub", Image: "icons/gcp/pubsub.png"},
		Lambda:       {Label: "Cloud Functions", Image: "icons/gcp/func.png"},
		LoadBalancer: {Label: "Load Balancer", Image: "icons/gcp/lb.png"},
		RDS:          {Label: "Big Table", Image: "icons/gcp/bigtable.png"},
		S3:           {Label: "GCS", Image: "icons/gcp/gcs.png"},
		SNS:          {Label: "SNS", Image: "icons/gcp/sns.png"},
		SQS:          {Label: "Pub Sub", Image: "icons/gcp/pubsub.png"},
		Question:     {Label: "TBD", Image: "icons/question.png"},
		BiqQuery:     {Label: "Biq Query", Image: "icons/gcp/bigquery.png"},
		Cloud:        {Label: "Cloud", Image: "icons/cloud.png"},
		Coralogix:    {Label: "Coralogix", Image: "icons/coralogix.svg"},
		Datadog:      {Label: "Datadog", Image: "icons/datadog.png"},
		Redis:        {Label: "Redis", Image: "icons/redis.png"},
	}
	legend := gcpMap[image]
	legendCh <- legend
	return legend.Image

}

func OrderLegend(in <-chan Legend) <-chan Legend {
	out := make(chan Legend)
	go func() {
		defer close(out)

		// Collect all legends from the input channel
		var legends []Legend
		for legend := range in {
			legends = append(legends, legend)
		}

		// Sort legends alphabetically by Label
		sort.Slice(legends, func(i, j int) bool {
			return legends[i].Label < legends[j].Label
		})

		// Deduplicate legends
		var seen = make(map[string]bool)
		for _, legend := range legends {
			if !seen[legend.Label] {
				seen[legend.Label] = true
				out <- legend
			}
		}
	}()
	return out
}
