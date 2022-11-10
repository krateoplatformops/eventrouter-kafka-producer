package eventrouter

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Shopify/sarama"
	kafka_sarama "github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/krateoplatformops/eventrouter-kafka-producer/internal/helpers/decode"
	"github.com/rs/zerolog"
)

func Handler(sender *kafka_sarama.Sender, verbose bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context()).With().Logger()
		if verbose {
			log = log.Level(zerolog.DebugLevel)
		}

		var nfo EventInfo
		err := decode.JSONBody(w, r, &nfo)
		if err != nil && !decode.IsEmptyBodyError(err) {
			log.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if verbose {
			log.Debug().Interface("event", nfo).Msg("Event received")
		}

		cli, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		if err != nil {
			log.Error().Err(err).Msg("Failed to create client.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sub, err := ToSFV(&nfo)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create SFV.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		e := cloudevents.NewEvent()
		e.SetID(nfo.Metadata.UID)
		e.SetType(fmt.Sprintf("%s.%s", nfo.Source, nfo.Reason))
		e.SetSource(fmt.Sprintf("%s.%s", nfo.Source, nfo.InvolvedObject.UID))
		e.SetSubject(sub)
		e.SetTime(nfo.Metadata.CreationTimestamp)
		e.SetExtension("partitionkey", nfo.DeploymentId)
		e.SetData(cloudevents.ApplicationJSON, nfo)

		res := cli.Send(
			// Set the producer message key
			kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())),
			e,
		)
		if cloudevents.IsUndelivered(res) {
			log.Error().Err(err).Msg("Failed to send message to Kafka.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if log.Debug().Enabled() {
			log.Debug().Str("eventId", e.ID()).Msgf("Message sent; accepted: %t", cloudevents.IsACK(res))
		}

		w.WriteHeader(http.StatusOK)
	})
}
