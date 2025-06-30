import React, { useState, useEffect } from 'react';
import {
  View,
  StyleSheet,
  FlatList,
  RefreshControl,
  Alert,
} from 'react-native';
import {
  Text,
  Card,
  Badge,
  ActivityIndicator,
  Button,
  Chip,
  Searchbar,
  Divider,
} from 'react-native-paper';
import { MaterialIcons } from '@expo/vector-icons';
import { roomGridAPI } from '../../src/services/api';
import { DailyBreakfastConsumption } from '../../src/types/roomgrid';

export default function HistoryScreen() {
  const [consumptionHistory, setConsumptionHistory] = useState<DailyBreakfastConsumption[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filteredHistory, setFilteredHistory] = useState<DailyBreakfastConsumption[]>([]);

  useEffect(() => {
    fetchConsumptionHistory();
  }, []);

  useEffect(() => {
    filterHistory();
  }, [searchQuery, consumptionHistory]);

  const fetchConsumptionHistory = async () => {
    try {
      const response = await roomGridAPI.getConsumptionHistory();
      const history = response.data.consumption_history || [];
      setConsumptionHistory(history);
    } catch (error) {
      Alert.alert('Error', 'Failed to load consumption history');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const filterHistory = () => {
    if (!searchQuery.trim()) {
      setFilteredHistory(consumptionHistory);
      return;
    }

    const filtered = consumptionHistory.filter(item => 
      item.room_number.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.consumption_date.includes(searchQuery)
    );
    setFilteredHistory(filtered);
  };

  const onRefresh = () => {
    setRefreshing(true);
    fetchConsumptionHistory();
  };

  const getStatusColor = (status: string) => {
    return status === 'consumed' ? '#2ECC71' : '#95A5A6';
  };

  const getStatusText = (status: string) => {
    return status === 'consumed' ? 'CONSUMED' : 'NOT CONSUMED';
  };

  const renderConsumptionItem = ({ item }: { item: DailyBreakfastConsumption }) => (
    <Card style={styles.consumptionCard}>
      <Card.Content>
        <View style={styles.consumptionHeader}>
          <View style={styles.roomInfo}>
            <Text variant="titleMedium">Room {item.room_number}</Text>
          </View>
          <Badge 
            style={[styles.statusBadge, { backgroundColor: getStatusColor(item.status) }]}
          >
            {getStatusText(item.status)}
          </Badge>
        </View>

        <Divider style={styles.divider} />

        <View style={styles.consumptionInfo}>
          <View style={styles.infoRow}>
            <MaterialIcons name="date-range" size={16} color="#666" />
            <Text variant="bodyMedium" style={styles.infoText}>
              {new Date(item.consumption_date).toLocaleDateString('en-US', {
                weekday: 'long',
                year: 'numeric',
                month: 'long',
                day: 'numeric'
              })}
            </Text>
          </View>

          {item.status === 'consumed' && item.consumed_at && (
            <View style={styles.infoRow}>
              <MaterialIcons name="access-time" size={16} color="#666" />
              <Text variant="bodyMedium" style={styles.infoText}>
                Consumed at {new Date(item.consumed_at).toLocaleTimeString('en-US', {
                  hour: '2-digit',
                  minute: '2-digit'
                })}
              </Text>
            </View>
          )}

          {item.pms_posted && (
            <View style={styles.infoRow}>
              <MaterialIcons name="receipt" size={16} color="#4ECDC4" />
              <Text variant="bodyMedium" style={[styles.infoText, { color: '#4ECDC4' }]}>
                Charged to PMS
              </Text>
            </View>
          )}

          {item.ohip_covered && (
            <Chip icon="medical-bag" style={styles.ohipChip}>
              OHIP Covered
            </Chip>
          )}
        </View>

        {item.notes && (
          <View style={styles.notesSection}>
            <Text variant="bodySmall" style={styles.notesLabel}>Notes:</Text>
            <Text variant="bodySmall" style={styles.notesText}>{item.notes}</Text>
          </View>
        )}
      </Card.Content>
    </Card>
  );

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" />
        <Text style={styles.loadingText}>Loading consumption history...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Searchbar
        placeholder="Search by room, guest, or date..."
        onChangeText={setSearchQuery}
        value={searchQuery}
        style={styles.searchBar}
      />
      
      <FlatList
        data={filteredHistory}
        renderItem={renderConsumptionItem}
        keyExtractor={(item) => `${item.id}`}
        contentContainerStyle={styles.historyList}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListEmptyComponent={() => (
          <View style={styles.emptyContainer}>
            <MaterialIcons name="history" size={64} color="#ccc" />
            <Text variant="titleLarge" style={styles.emptyTitle}>
              No Consumption History
            </Text>
            <Text variant="bodyMedium" style={styles.emptyText}>
              Breakfast consumption records will appear here once guests start consuming their breakfast packages.
            </Text>
          </View>
        )}
      />
    </View>
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
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
  },
  searchBar: {
    margin: 16,
    marginBottom: 8,
  },
  historyList: {
    padding: 16,
    paddingTop: 8,
  },
  consumptionCard: {
    marginBottom: 16,
    elevation: 4,
  },
  consumptionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  roomInfo: {
    flex: 1,
  },
  guestName: {
    color: '#666',
    marginTop: 2,
  },
  statusBadge: {
    color: 'white',
  },
  divider: {
    marginBottom: 12,
  },
  consumptionInfo: {
    marginBottom: 12,
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 6,
  },
  infoText: {
    marginLeft: 8,
    color: '#666',
  },
  ohipChip: {
    backgroundColor: '#E8F8F5',
    alignSelf: 'flex-start',
    marginTop: 8,
  },
  notesSection: {
    backgroundColor: '#f8f9fa',
    padding: 12,
    borderRadius: 8,
    marginTop: 8,
  },
  notesLabel: {
    fontWeight: 'bold',
    marginBottom: 4,
    color: '#333',
  },
  notesText: {
    color: '#666',
    fontStyle: 'italic',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 64,
  },
  emptyTitle: {
    marginTop: 16,
    marginBottom: 8,
    color: '#666',
  },
  emptyText: {
    color: '#999',
    textAlign: 'center',
    paddingHorizontal: 32,
  },
});
