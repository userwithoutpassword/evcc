package charger

// LICENSE

// Copyright (c) 2024 andig

// This module is NOT covered by the MIT license. All rights reserved.

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"context"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/charger/heating"
	"github.com/evcc-io/evcc/plugin"
	"github.com/evcc-io/evcc/util"
)

// SgReady charger implementation
type SgReady struct {
	*embed
	*heating.SgReadyModeController
	*heating.PowerController
}

func init() {
	registry.AddCtx("sgready", NewSgReadyFromConfig)
}

//go:generate go tool decorate -f decorateSgReady -b *SgReady -r api.Charger -t "api.Meter,CurrentPower,func() (float64, error)" -t "api.MeterEnergy,TotalEnergy,func() (float64, error)" -t "api.Battery,Soc,func() (float64, error)" -t "api.SocLimiter,GetLimitSoc,func() (int64, error)"

// NewSgReadyFromConfig creates an SG Ready configurable charger from generic config
func NewSgReadyFromConfig(ctx context.Context, other map[string]interface{}) (api.Charger, error) {
	cc := struct {
		embed            `mapstructure:",squash"`
		SetMode          plugin.Config
		GetMode          *plugin.Config // optional
		SetMaxPower      *plugin.Config // optional
		heating.Readings `mapstructure:",squash"`
		Phases           int
	}{
		embed: embed{
			Icon_:     "heatpump",
			Features_: []api.Feature{api.Heating, api.IntegratedDevice},
		},
		Phases: 1,
	}

	if err := util.DecodeOther(other, &cc); err != nil {
		return nil, err
	}

	modeS, err := cc.SetMode.IntSetter(ctx, "mode")
	if err != nil {
		return nil, err
	}

	modeG, err := cc.GetMode.IntGetter(ctx)
	if err != nil {
		return nil, err
	}

	maxPowerS, err := cc.SetMaxPower.IntSetter(ctx, "maxpower")
	if err != nil {
		return nil, err
	}

	res, err := NewSgReady(ctx, &cc.embed, modeS, modeG, maxPowerS, cc.Phases)
	if err != nil {
		return nil, err
	}

	powerG, energyG, tempG, limitTempG, err := cc.Readings.Configure(ctx)
	if err != nil {
		return nil, err
	}

	return decorateSgReady(res, powerG, energyG, tempG, limitTempG), nil
}

// NewSgReady creates SG Ready charger
func NewSgReady(ctx context.Context, embed *embed, modeS func(int64) error, modeG func() (int64, error), maxPowerS func(int64) error, phases int) (*SgReady, error) {
	res := &SgReady{
		embed:                 embed,
		SgReadyModeController: heating.NewSgReadyModeController(ctx, modeS, modeG),
		PowerController:       heating.NewPowerController(ctx, maxPowerS, phases),
	}

	return res, nil
}
