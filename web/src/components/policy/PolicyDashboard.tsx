import { useCallback, useEffect, useState } from 'react'
import {
  AlertTriangle,
  CheckCircle,
  FileText,
  RefreshCw,
  Shield,
  TrendingUp,
  XCircle,
} from 'lucide-react'

interface PolicyViolation {
  policy: string
  severity: string
  message: string
  path?: string
}

interface PolicyReport {
  timestamp: string
  policySet: string
  enforcement: string
  passed: number
  failed: number
  warnings: number
  violations: PolicyViolation[]
}

interface PolicyDashboardProps {
  apiEndpoint?: string
}

function severityPillClass(severity: string): string {
  switch (severity.toLowerCase()) {
    case 'critical':
      return 'bg-red-500 text-white'
    case 'error':
      return 'bg-orange-500 text-white'
    case 'warning':
      return 'bg-yellow-500 text-black'
    case 'info':
      return 'bg-blue-500 text-white'
    default:
      return 'bg-gray-500 text-white'
  }
}

export function PolicyDashboard({ apiEndpoint = '/api/v1/policy' }: PolicyDashboardProps) {
  const [report, setReport] = useState<PolicyReport | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchPolicyReport = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`${apiEndpoint}/report`)
      if (!response.ok) {
        throw new Error('Failed to fetch policy report')
      }

      const data = (await response.json()) as PolicyReport
      setReport(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }, [apiEndpoint])

  useEffect(() => {
    void fetchPolicyReport()
  }, [fetchPolicyReport])

  const getStatusIcon = (passed: number, failed: number) => {
    if (failed > 0) {
      return <XCircle className="w-5 h-5 text-red-500" />
    }
    if (passed > 0) {
      return <CheckCircle className="w-5 h-5 text-green-500" />
    }
    return <AlertTriangle className="w-5 h-5 text-yellow-500" />
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    )
  }

  if (error) {
    return (
      <div
        className="flex items-start gap-2 rounded-md border border-red-200 bg-red-50 p-4 text-red-800"
        role="alert"
      >
        <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5" />
        <p className="text-sm">{error}</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="w-8 h-8 text-blue-500" />
          <div>
            <h1 className="text-2xl font-bold">Policy Compliance</h1>
            <p className="text-gray-500">Real-time policy evaluation results</p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => void fetchPolicyReport()}
          className="inline-flex items-center gap-2 rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium shadow-sm hover:bg-gray-50"
        >
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
        {[
          {
            title: 'Total Policies',
            icon: <FileText className="h-4 w-4 text-gray-500" />,
            value: (report?.passed || 0) + (report?.failed || 0),
            valueClass: '',
          },
          {
            title: 'Passed',
            icon: <CheckCircle className="h-4 w-4 text-green-500" />,
            value: report?.passed || 0,
            valueClass: 'text-green-600',
          },
          {
            title: 'Failed',
            icon: <XCircle className="h-4 w-4 text-red-500" />,
            value: report?.failed || 0,
            valueClass: 'text-red-600',
          },
          {
            title: 'Warnings',
            icon: <AlertTriangle className="h-4 w-4 text-yellow-500" />,
            value: report?.warnings || 0,
            valueClass: 'text-yellow-600',
          },
        ].map((card) => (
          <div key={card.title} className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
            <div className="flex flex-row items-center justify-between pb-2">
              <span className="text-sm font-medium text-gray-700">{card.title}</span>
              {card.icon}
            </div>
            <div className={`text-2xl font-bold ${card.valueClass}`}>{card.value}</div>
          </div>
        ))}
      </div>

      <div className="rounded-lg border border-gray-200 bg-white shadow-sm">
        <div className="border-b border-gray-100 p-4">
          <h2 className="flex items-center gap-2 text-lg font-semibold">
            {getStatusIcon(report?.passed || 0, report?.failed || 0)}
            Compliance Status
          </h2>
        </div>
        <div className="p-4">
          <div className="flex items-center gap-4">
            <div className="flex-1">
              <div className="mb-1 flex justify-between">
                <span className="text-sm font-medium">Overall Compliance</span>
                <span className="text-sm text-gray-500">
                  {report?.passed || 0} / {(report?.passed || 0) + (report?.failed || 0)}
                </span>
              </div>
              <div className="h-2.5 w-full rounded-full bg-gray-200">
                <div
                  className="h-2.5 rounded-full bg-green-500"
                  style={{
                    width: `${((report?.passed || 0) / ((report?.passed || 0) + (report?.failed || 0) || 1)) * 100}%`,
                  }}
                />
              </div>
            </div>
          </div>

          {report?.timestamp && (
            <p className="mt-4 text-sm text-gray-500">
              Last evaluated: {new Date(report.timestamp).toLocaleString()}
            </p>
          )}
        </div>
      </div>

      {report?.violations && report.violations.length > 0 && (
        <div className="rounded-lg border border-gray-200 bg-white shadow-sm">
          <div className="border-b border-gray-100 p-4">
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              Policy Violations ({report.violations.length})
            </h2>
          </div>
          <div className="space-y-3 p-4">
            {report.violations.map((violation, index) => (
              <div key={index} className="flex items-start gap-3 rounded-lg bg-gray-50 p-3">
                <span
                  className={`inline-flex shrink-0 rounded px-2 py-0.5 text-xs font-semibold ${severityPillClass(violation.severity)}`}
                >
                  {violation.severity.toUpperCase()}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="font-medium">{violation.policy}</p>
                  <p className="text-sm text-gray-600">{violation.message}</p>
                  {violation.path && (
                    <p className="mt-1 text-xs text-gray-400">Path: {violation.path}</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="rounded-lg border border-gray-200 bg-white shadow-sm">
        <div className="border-b border-gray-100 p-4">
          <h2 className="flex items-center gap-2 text-lg font-semibold">
            <TrendingUp className="w-5 h-5" />
            Compliance Trend
          </h2>
        </div>
        <div className="flex h-64 flex-col items-center justify-center text-gray-400">
          <p>Trend visualization would be implemented here</p>
          <p className="text-sm">(Chart.js or similar library)</p>
        </div>
      </div>
    </div>
  )
}
