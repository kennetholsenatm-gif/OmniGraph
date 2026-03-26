import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Alert, AlertDescription } from '../ui/alert';
import { 
  Shield, 
  AlertTriangle, 
  CheckCircle, 
  XCircle, 
  RefreshCw,
  FileText,
  TrendingUp
} from 'lucide-react';

interface PolicyViolation {
  policy: string;
  severity: string;
  message: string;
  path?: string;
}

interface PolicyReport {
  timestamp: string;
  policySet: string;
  enforcement: string;
  passed: number;
  failed: number;
  warnings: number;
  violations: PolicyViolation[];
}

interface PolicyDashboardProps {
  apiEndpoint?: string;
}

export function PolicyDashboard({ apiEndpoint = '/api/v1/policy' }: PolicyDashboardProps) {
  const [report, setReport] = useState<PolicyReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPolicyReport();
  }, []);

  const fetchPolicyReport = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`${apiEndpoint}/report`);
      if (!response.ok) {
        throw new Error('Failed to fetch policy report');
      }
      
      const data = await response.json();
      setReport(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return 'bg-red-500';
      case 'error':
        return 'bg-orange-500';
      case 'warning':
        return 'bg-yellow-500';
      case 'info':
        return 'bg-blue-500';
      default:
        return 'bg-gray-500';
    }
  };

  const getStatusIcon = (passed: number, failed: number) => {
    if (failed > 0) {
      return <XCircle className="w-5 h-5 text-red-500" />;
    }
    if (passed > 0) {
      return <CheckCircle className="w-5 h-5 text-green-500" />;
    }
    return <AlertTriangle className="w-5 h-5 text-yellow-500" />;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertTriangle className="h-4 w-4" />
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="w-8 h-8 text-blue-500" />
          <div>
            <h1 className="text-2xl font-bold">Policy Compliance</h1>
            <p className="text-gray-500">Real-time policy evaluation results</p>
          </div>
        </div>
        <Button onClick={fetchPolicyReport} variant="outline" size="sm">
          <RefreshCw className="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Policies</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(report?.passed || 0) + (report?.failed || 0)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Passed</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {report?.passed || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Failed</CardTitle>
            <XCircle className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {report?.failed || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Warnings</CardTitle>
            <AlertTriangle className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">
              {report?.warnings || 0}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Status Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {getStatusIcon(report?.passed || 0, report?.failed || 0)}
            Compliance Status
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className="flex-1">
              <div className="flex justify-between mb-1">
                <span className="text-sm font-medium">Overall Compliance</span>
                <span className="text-sm text-gray-500">
                  {report?.passed || 0} / {(report?.passed || 0) + (report?.failed || 0)}
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div 
                  className="bg-green-500 h-2.5 rounded-full" 
                  style={{ 
                    width: `${((report?.passed || 0) / ((report?.passed || 0) + (report?.failed || 0) || 1)) * 100}%` 
                  }}
                ></div>
              </div>
            </div>
          </div>
          
          {report?.timestamp && (
            <p className="text-sm text-gray-500 mt-4">
              Last evaluated: {new Date(report.timestamp).toLocaleString()}
            </p>
          )}
        </CardContent>
      </Card>

      {/* Violations List */}
      {report?.violations && report.violations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              Policy Violations ({report.violations.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {report.violations.map((violation, index) => (
                <div 
                  key={index}
                  className="flex items-start gap-3 p-3 bg-gray-50 rounded-lg"
                >
                  <Badge className={getSeverityColor(violation.severity)}>
                    {violation.severity.toUpperCase()}
                  </Badge>
                  <div className="flex-1">
                    <p className="font-medium">{violation.policy}</p>
                    <p className="text-sm text-gray-600">{violation.message}</p>
                    {violation.path && (
                      <p className="text-xs text-gray-400 mt-1">
                        Path: {violation.path}
                      </p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Trend Chart Placeholder */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="w-5 h-5" />
            Compliance Trend
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center text-gray-400">
            <p>Trend visualization would be implemented here</p>
            <p className="text-sm">(Chart.js or similar library)</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}