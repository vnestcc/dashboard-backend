package company

import (
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/vnestcc/dashboard/models"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

type Claims = middleware.Claims

var StartupCache = cacher.NewCacher[uint, models.Company](&cacher.NewCacherOpts{
	TimeToLive:    2 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
	CleanerMode:   cacher.CleaningCentral,
})

var QuarterCache = cacher.NewCacher[string, models.Quarter](&cacher.NewCacherOpts{
	TimeToLive:    3 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
})

type createCompanyRequest struct {
	Name         string `json:"name" binding:"required" example:"Acme Inc"`
	ContactName  string `json:"contact_name" binding:"required" example:"John Doe"`
	ContactEmail string `json:"contact_email" binding:"required,email" example:"john@acme.com"`
	Sector       string `json:"sector" binding:"required" example:"xyz"`
	Description  string `json:"description" binding:"required" example:"We do something xyz and make money"`
}

type nextQuarter struct {
	NextQuarter string `json:"next_quarter" binding:"required" example:"Q1"`
	NextYear    uint   `json:"next_year" binding:"required" example:"2025"`
}

type quarterResponse struct {
	ID      uint   `json:"id" example:"1"`
	Quarter string `json:"quarter" example:"Q1"`
	Year    uint   `json:"year" example:"2025"`
	Date    string `json:"date,omitempty" example:"2025-04-01T00:00:00Z"`
}

type joinCompanyRequest struct {
	SecretCode string `json:"secret_code" binding:"required" example:"random hex"`
}

type quarterRequest struct {
	Quarter string `json:"quarter" binding:"required" example:"Q1"`
	Year    uint   `json:"year" binding:"required" example:"2024"`
}

type versionInfo struct {
	Version    uint32
	IsEditable uint16
}

type editableModel interface {
	EditableFilter() error
}

type financeMetric struct {
	QuarterlyRevenue string    `json:"quarterly_revenue"`
	RevenueGrowth    string    `json:"revenue_growth"`
	GrossMargin      string    `json:"gross_margin"`
	NetMargin        string    `json:"net_margin"`
	Quarter          string    `json:"quarter"`
	Year             uint      `json:"year"`
	Date             time.Time `json:"date"`
}

type marketMetric struct {
	TotalCustomers string `json:"total_customers"`
	CustomerGrowth string `json:"customer_growth"`
	ConversionRate string `json:"conversion_rate"`
	RetentionRate  string `json:"retention_rate"`
	ChurnRate      string `json:"churn_rate"`
	Quarter        string `json:"quarter"`
	Year           string `json:"year"`
	Date           string `json:"date"`
}

type economicsMetric struct {
	CAC        string `json:"cac"`
	CACPayback string `json:"cac_payback"`
	ARPU       string `json:"arpu"`
	LTV        string `json:"ltv"`
	Quarter    string `json:"quarter"`
	Year       string `json:"year"`
	Date       string `json:"date"`
}

type productMetric struct {
	ActiveUsers         string `json:"active_users"`
	EngagementMetrics   string `json:"engagement_metrics"`
	MilestonesAchieved  string `json:"milestones_achieved"`
	MilestonesMissed    string `json:"milestones_missed"`
	Roadmap             string `json:"roadmap"`
	TechnicalChallenges string `json:"technical_challenges"`
	ProductBottlenecks  string `json:"product_bottlenecks"`
	Quarter             string `json:"quarter"`
	Year                string `json:"year"`
	Date                string `json:"date"`
}

type teamperfMetric struct {
	TeamStrengths          string `json:"team_strengths"`
	DevelopmentInitiatives string `json:"development_initiatives"`
	TeamSize               string `json:"team_size"`
	NewHires               string `json:"new_hires"`
	Turnover               string `json:"turnover"`
	VacantPositions        string `json:"vacant_positions"`
	LeadershipAlignment    string `json:"leadership_alignment"`
	SkillGaps              string `json:"skill_gaps"`
	Quarter                string `json:"quarter"`
	Year                   string `json:"year"`
	Date                   string `json:"date"`
}

type fundMetric struct {
	LastRound             string `json:"last_round"`
	TargetAmount          string `json:"target_amount"`
	ValuationExpectations string `json:"valuation_expectations"`
	Quarter               string `json:"quarter"`
	Year                  string `json:"year"`
	Date                  string `json:"date"`
}

type operationalMetric struct {
	InfrastructureCapacity string `json:"infrastructure_capacity"`
	OperationalBottlenecks string `json:"operational_bottlenecks"`
	OptimizationAreas      string `json:"optimization_areas"`
	ScalingPlans           string `json:"scaling_plans"`
	Quarter                string `json:"quarter"`
	Year                   string `json:"year"`
	Date                   string `json:"date"`
}

type riskMetric struct {
	ComplianceStatus   string `json:"compliance_status"`
	SecurityIncidents  string `json:"security_incidents"`
	RegulatoryConcerns string `json:"regulatory_concerns"`
	Quarter            string `json:"quarter"`
	Year               string `json:"year"`
	Date               string `json:"date"`
}

type additionalMetric struct {
	InitiativeProgress       string `json:"initiative_progress"`
	GrowthChallenges         string `json:"growth_challenges"`
	BusinessModelAdjustments string `json:"business_model_adjustments"`
	Quarter                  string `json:"quarter"`
	Year                     string `json:"year"`
	Date                     string `json:"date"`
}

type assessmentMetric struct {
	FinancialRating uint8  `json:"finance_rating"`
	MarketRating    uint8  `json:"market_rating"`
	OverallRating   uint8  `json:"overall_rating"`
	Quarter         string `json:"quarter"`
	Year            string `json:"year"`
	Date            string `json:"date"`
}
