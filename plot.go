package slackActivity

import (
	"time"

	"github.com/slack-go/slack"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func GeneratePlot(counts []MessageCount, channel slack.Channel, imageHeight int, imageWidth int, outputPath string) error {
	values := make(plotter.Values, len(counts))
	for i, r := range counts {
		values[i] = float64(r.Count)
	}
	p := plot.New()
	p.Title.Text = channel.Name + "[reported at " + time.Now().Format("2006/01/02") + " ]"
	barwidth := float64(imageWidth / len(counts))
	w := vg.Points(barwidth)
	bars, err := plotter.NewBarChart(values, w)
	if err != nil {
		return err
	}
	bars.LineStyle.Width = vg.Length(0)
	bars.Color = plotutil.Color(1)
	p.Add(bars)
	if err := p.Save(vg.Length(imageWidth), vg.Length(imageHeight), outputPath); err != nil {
		return err
	}
	return nil
}
