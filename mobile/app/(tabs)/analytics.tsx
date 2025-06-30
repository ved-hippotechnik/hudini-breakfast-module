import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  RefreshControl,
  Alert,
} from 'react-native';
import { Card, Button, ActivityIndicator } from 'react-native-paper';
import { MaterialIcons } from '@expo/vector-icons';
import { useAuth } from '../../src/contexts/AuthContext';
import { analyticsAPI } from '../../src/services/api';

interface BreakfastReportData {
  totalRoomsWithBreakfast: number;
  consumptionRate: number;
  totalConsumed: number;
  totalNotConsumed: number;
  averageConsumptionTime: string;
  ohipCoveredCount: number;
  pmsChargesPosted: number;
  dailyTrend: Array<{
    date: string;
    consumed: number;
    total: number;
    rate: number;
  }>;
  roomTypeBreakdown: Array<{
    roomType: string;
    totalRooms: number;
    consumed: number;
    rate: number;
  }>;
}

export default function ReportsScreen() {
  const { user } = useAuth();
  const [reportData, setReportData] = useState<BreakfastReportData | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [selectedPeriod, setSelectedPeriod] = useState<'today' | 'week' | 'month'>('today');

  useEffect(() => {
    fetchReportData();
  }, [selectedPeriod]);

  const fetchReportData = async () => {
    try {
      setLoading(true);
      
      // Simulated data for demonstration
      const simulatedData: BreakfastReportData = {
        totalRoomsWithBreakfast: selectedPeriod === 'today' ? 45 : selectedPeriod === 'week' ? 315 : 1350,
        consumptionRate: selectedPeriod === 'today' ? 78.3 : selectedPeriod === 'week' ? 82.1 : 79.8,
        totalConsumed: selectedPeriod === 'today' ? 35 : selectedPeriod === 'week' ? 259 : 1077,
        totalNotConsumed: selectedPeriod === 'today' ? 10 : selectedPeriod === 'week' ? 56 : 273,
        averageConsumptionTime: selectedPeriod === 'today' ? '8:45 AM' : selectedPeriod === 'week' ? '8:30 AM' : '8:35 AM',
        ohipCoveredCount: selectedPeriod === 'today' ? 12 : selectedPeriod === 'week' ? 89 : 387,
        pmsChargesPosted: selectedPeriod === 'today' ? 33 : selectedPeriod === 'week' ? 251 : 1045,
        dailyTrend: generateTrendData(selectedPeriod),
        roomTypeBreakdown: [
          { roomType: 'Standard', totalRooms: selectedPeriod === 'today' ? 25 : 175, consumed: selectedPeriod === 'today' ? 19 : 142, rate: 76.0 },
          { roomType: 'Deluxe', totalRooms: selectedPeriod === 'today' ? 15 : 105, consumed: selectedPeriod === 'today' ? 13 : 89, rate: 84.8 },
          { roomType: 'Suite', totalRooms: selectedPeriod === 'today' ? 5 : 35, consumed: selectedPeriod === 'today' ? 3 : 28, rate: 80.0 },
        ],
      };
      
      setReportData(simulatedData);
    } catch (error) {
      Alert.alert('Error', 'Failed to fetch report data');
      console.error('Error fetching report data:', error);
    } finally {
      setLoading(false);
    }
  };

  const generateTrendData = (period: string) => {
    if (period === 'today') {
      return [
        { date: 'Today', consumed: 35, total: 45, rate: 77.8 },
      ];
    } else if (period === 'week') {
      return [
        { date: 'Mon', consumed: 38, total: 45, rate: 84.4 },
        { date: 'Tue', consumed: 42, total: 47, rate: 89.4 },
        { date: 'Wed', consumed: 35, total: 43, rate: 81.4 },
        { date: 'Thu', consumed: 40, total: 46, rate: 87.0 },
        { date: 'Fri', consumed: 39, total: 48, rate: 81.3 },
        { date: 'Sat', consumed: 33, total: 41, rate: 80.5 },
        { date: 'Sun', consumed: 32, total: 45, rate: 71.1 },
      ];
    } else {
      return [
        { date: 'Week 1', consumed: 278, total: 315, rate: 88.3 },
        { date: 'Week 2', consumed: 256, total: 320, rate: 80.0 },
        { date: 'Week 3', consumed: 289, total: 335, rate: 86.3 },
        { date: 'Week 4', consumed: 254, total: 330, rate: 77.0 },
      ];
    }
  };

  const onRefresh = async () => {
    setRefreshing(true);
    await fetchReportData();
    setRefreshing(false);
  };

  const formatPercentage = (value: number) => `${value.toFixed(1)}%`;
  const formatNumber = (value: number) => value.toLocaleString();

  if (loading && !reportData) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" />
        <Text style={styles.loadingText}>Loading breakfast reports...</Text>
      </View>
    );
  }

  return (
    <ScrollView 
      style={styles.container}
      refreshControl={
        <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
      }
    >
      <View style={styles.header}>
        <Text style={styles.title}>Breakfast Reports</Text>
        <View style={styles.periodSelector}>
          <Button
            mode={selectedPeriod === 'today' ? 'contained' : 'outlined'}
            compact
            onPress={() => setSelectedPeriod('today')}
            style={styles.periodButton}
          >
            Today
          </Button>
          <Button
            mode={selectedPeriod === 'week' ? 'contained' : 'outlined'}
            compact
            onPress={() => setSelectedPeriod('week')}
            style={styles.periodButton}
          >
            Week
          </Button>
          <Button
            mode={selectedPeriod === 'month' ? 'contained' : 'outlined'}
            compact
            onPress={() => setSelectedPeriod('month')}
            style={styles.periodButton}
          >
            Month
          </Button>
        </View>
      </View>
      
      {/* Key Metrics */}
      <View style={styles.metricsContainer}>
        <Card style={styles.metricCard}>
          <Card.Content style={styles.metricContent}>
            <MaterialIcons name="hotel" size={24} color="#4ECDC4" />
            <Text style={styles.metricNumber}>{formatNumber(reportData?.totalRoomsWithBreakfast || 0)}</Text>
            <Text style={styles.metricLabel}>Rooms w/ Breakfast</Text>
          </Card.Content>
        </Card>
        
        <Card style={styles.metricCard}>
          <Card.Content style={styles.metricContent}>
            <MaterialIcons name="trending-up" size={24} color="#2ECC71" />
            <Text style={styles.metricNumber}>{formatPercentage(reportData?.consumptionRate || 0)}</Text>
            <Text style={styles.metricLabel}>Consumption Rate</Text>
          </Card.Content>
        </Card>
      </View>

      <View style={styles.metricsContainer}>
        <Card style={styles.metricCard}>
          <Card.Content style={styles.metricContent}>
            <MaterialIcons name="check-circle" size={24} color="#27AE60" />
            <Text style={styles.metricNumber}>{formatNumber(reportData?.totalConsumed || 0)}</Text>
            <Text style={styles.metricLabel}>Consumed</Text>
          </Card.Content>
        </Card>
        
        <Card style={styles.metricCard}>
          <Card.Content style={styles.metricContent}>
            <MaterialIcons name="cancel" size={24} color="#E74C3C" />
            <Text style={styles.metricNumber}>{formatNumber(reportData?.totalNotConsumed || 0)}</Text>
            <Text style={styles.metricLabel}>Not Consumed</Text>
          </Card.Content>
        </Card>
      </View>

      {/* Consumption Trend */}
      <Card style={styles.card}>
        <Card.Title title="Consumption Trend" />
        <Card.Content>
          <ScrollView horizontal showsHorizontalScrollIndicator={false}>
            <View style={styles.trendChart}>
              {reportData?.dailyTrend.map((day, index) => (
                <View key={index} style={styles.trendDay}>
                  <View 
                    style={[
                      styles.trendBar,
                      { height: (day.rate / 100) * 80 }
                    ]}
                  />
                  <Text style={styles.trendRate}>{formatPercentage(day.rate)}</Text>
                  <Text style={styles.trendDate}>{day.date}</Text>
                  <Text style={styles.trendCount}>{day.consumed}/{day.total}</Text>
                </View>
              ))}
            </View>
          </ScrollView>
        </Card.Content>
      </Card>

      {/* Room Type Breakdown */}
      <Card style={styles.card}>
        <Card.Title title="Consumption by Room Type" />
        <Card.Content>
          {reportData?.roomTypeBreakdown.map((roomType, index) => (
            <View key={index} style={styles.roomTypeItem}>
              <View style={styles.roomTypeHeader}>
                <Text style={styles.roomTypeName}>{roomType.roomType}</Text>
                <Text style={styles.roomTypeRate}>{formatPercentage(roomType.rate)}</Text>
              </View>
              <View style={styles.roomTypeStats}>
                <Text style={styles.roomTypeNumbers}>
                  {roomType.consumed} consumed of {roomType.totalRooms} rooms
                </Text>
              </View>
              <View style={styles.roomTypeBar}>
                <View 
                  style={[
                    styles.roomTypeBarFill,
                    { width: `${roomType.rate}%` }
                  ]}
                />
              </View>
            </View>
          ))}
        </Card.Content>
      </Card>

      {/* Additional Metrics */}
      <View style={styles.metricsContainer}>
        <Card style={styles.metricCard}>
          <Card.Content style={styles.metricContent}>
            <MaterialIcons name="access-time" size={24} color="#F39C12" />
            <Text style={styles.metricNumber}>{reportData?.averageConsumptionTime || 'N/A'}</Text>
            <Text style={styles.metricLabel}>Avg. Time</Text>
          </Card.Content>
        </Card>
        
        <Card style={styles.metricCard}>
          <Card.Content style={styles.metricContent}>
            <MaterialIcons name="local-hospital" size={24} color="#9B59B6" />
            <Text style={styles.metricNumber}>{formatNumber(reportData?.ohipCoveredCount || 0)}</Text>
            <Text style={styles.metricLabel}>OHIP Covered</Text>
          </Card.Content>
        </Card>
      </View>

      <Card style={styles.card}>
        <Card.Title title="PMS Integration" />
        <Card.Content>
          <View style={styles.pmsStats}>
            <MaterialIcons name="receipt" size={24} color="#3498DB" />
            <Text style={styles.pmsNumber}>{formatNumber(reportData?.pmsChargesPosted || 0)}</Text>
            <Text style={styles.pmsLabel}>Charges Posted to PMS</Text>
          </View>
        </Card.Content>
      </Card>

      {/* Insights */}
      <Card style={styles.card}>
        <Card.Title title="Insights" />
        <Card.Content>
          <View style={styles.insight}>
            <MaterialIcons name="lightbulb-outline" size={20} color="#F39C12" />
            <Text style={styles.insightText}>
              {selectedPeriod === 'today' 
                ? "Today's consumption rate is 78.3%, slightly below the weekly average of 82.1%."
                : selectedPeriod === 'week'
                ? "This week's consumption rate (82.1%) is above the monthly average of 79.8%."
                : "Monthly consumption rate is steady at 79.8% with consistent performance across room types."
              }
            </Text>
          </View>
          <View style={styles.insight}>
            <MaterialIcons name="trending-up" size={20} color="#2ECC71" />
            <Text style={styles.insightText}>
              Deluxe rooms show the highest consumption rate across all periods.
            </Text>
          </View>
          <View style={styles.insight}>
            <MaterialIcons name="schedule" size={20} color="#E67E22" />
            <Text style={styles.insightText}>
              Peak consumption time is between 8:30-8:45 AM across all room types.
            </Text>
          </View>
        </Card.Content>
      </Card>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f5f5f5',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#666',
  },
  header: {
    padding: 20,
    backgroundColor: '#fff',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#2C3E50',
    marginBottom: 16,
  },
  periodSelector: {
    flexDirection: 'row',
    gap: 8,
  },
  periodButton: {
    marginRight: 8,
  },
  metricsContainer: {
    flexDirection: 'row',
    paddingHorizontal: 20,
    paddingTop: 16,
    gap: 16,
  },
  metricCard: {
    flex: 1,
    backgroundColor: '#fff',
  },
  metricContent: {
    alignItems: 'center',
    paddingVertical: 16,
  },
  metricNumber: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#2C3E50',
    marginTop: 8,
  },
  metricLabel: {
    fontSize: 12,
    color: '#7F8C8D',
    textAlign: 'center',
    marginTop: 4,
  },
  card: {
    margin: 20,
    marginTop: 16,
    backgroundColor: '#fff',
  },
  trendChart: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    paddingHorizontal: 16,
    paddingVertical: 20,
  },
  trendDay: {
    alignItems: 'center',
    marginHorizontal: 8,
    minWidth: 50,
  },
  trendBar: {
    width: 20,
    backgroundColor: '#3498DB',
    marginBottom: 8,
    borderRadius: 10,
  },
  trendRate: {
    fontSize: 10,
    fontWeight: 'bold',
    color: '#2C3E50',
    marginBottom: 4,
  },
  trendDate: {
    fontSize: 10,
    color: '#7F8C8D',
    marginBottom: 2,
  },
  trendCount: {
    fontSize: 9,
    color: '#95A5A6',
  },
  roomTypeItem: {
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#ECF0F1',
  },
  roomTypeHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  roomTypeName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#2C3E50',
  },
  roomTypeRate: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#27AE60',
  },
  roomTypeStats: {
    marginBottom: 8,
  },
  roomTypeNumbers: {
    fontSize: 12,
    color: '#7F8C8D',
  },
  roomTypeBar: {
    height: 6,
    backgroundColor: '#ECF0F1',
    borderRadius: 3,
    overflow: 'hidden',
  },
  roomTypeBarFill: {
    height: '100%',
    backgroundColor: '#27AE60',
    borderRadius: 3,
  },
  pmsStats: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
  },
  pmsNumber: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#3498DB',
    marginLeft: 12,
    marginRight: 8,
  },
  pmsLabel: {
    fontSize: 14,
    color: '#7F8C8D',
    flex: 1,
  },
  insight: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    paddingVertical: 8,
  },
  insightText: {
    fontSize: 14,
    color: '#34495E',
    marginLeft: 12,
    flex: 1,
    lineHeight: 20,
  },
});
