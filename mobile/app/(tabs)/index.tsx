import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  View,
  StyleSheet,
  FlatList,
  Alert,
  RefreshControl,
  ScrollView,
  Dimensions,
  TouchableOpacity,
} from 'react-native';
import {
  Text,
  Card,
  Button,
  FAB,
  Badge,
  ActivityIndicator,
  Searchbar,
  Chip,
  Menu,
  Portal,
  Modal,
  Surface,
  Divider,
} from 'react-native-paper';
import { roomGridAPI } from '../../src/services/api';
import { useAuth } from '../../src/contexts/AuthContext';
import { RoomBreakfastStatus, APIResponse, MarkConsumptionRequest } from '../../src/types/api';

const { width } = Dimensions.get('window');
const ROOM_CARD_SIZE = (width - 60) / 4; // 4 rooms per row with margins

export default function RoomGridScreen() {
  const [rooms, setRooms] = useState<RoomBreakfastStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [showFilterMenu, setShowFilterMenu] = useState(false);
  const [selectedDate, setSelectedDate] = useState(new Date().toISOString().split('T')[0]);
  
  const { user } = useAuth();
  const propertyId = user?.property_id || 'PROP001'; // Default for demo

  const statusFilters = [
    { value: 'all', label: 'All Rooms' },
    { value: 'has_breakfast', label: 'Has Breakfast' },
    { value: 'consumed', label: 'Consumed Today' },
    { value: 'pending', label: 'Pending Consumption' },
    { value: 'no_guest', label: 'No Guest' },
  ];

  useEffect(() => {
    fetchRoomGrid();
  }, [fetchRoomGrid]);

  const fetchRoomGrid = useCallback(async () => {
    try {
      setLoading(true);
      const response = await roomGridAPI.getRoomGrid(propertyId, selectedDate);
      setRooms(response.data.rooms || []);
    } catch (error) {
      console.error('Error fetching room grid:', error);
      Alert.alert('Error', 'Failed to load room grid data');
    } finally {
      setLoading(false);
    }
  }, [propertyId, selectedDate]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchRoomGrid();
    setRefreshing(false);
  }, [fetchRoomGrid]);

  const syncFromPMS = useCallback(async () => {
    try {
      setLoading(true);
      await roomGridAPI.syncFromPMS(propertyId);
      await fetchRoomGrid();
      Alert.alert('Success', 'Guest data synchronized from PMS');
    } catch (error) {
      console.error('Error syncing from PMS:', error);
      Alert.alert('Error', 'Failed to sync from PMS');
    } finally {
      setLoading(false);
    }
  }, [propertyId, fetchRoomGrid]);

  const markBreakfastConsumed = useCallback(async (roomNumber: string) => {
    try {
      const data: MarkConsumptionRequest = {
        property_id: propertyId,
        room_number: roomNumber,
        payment_method: 'room_charge', // Default, could be configurable
        notes: `Consumed via mobile app by ${user?.first_name} ${user?.last_name}`,
      };
      
      await roomGridAPI.markBreakfastConsumed(data);
      await fetchRoomGrid();
      Alert.alert('Success', `Breakfast marked as consumed for room ${roomNumber}`);
    } catch (error: any) {
      console.error('Error marking breakfast consumed:', error);
      const message = error.response?.data?.error?.message || 'Failed to mark breakfast as consumed';
      Alert.alert('Error', message);
    }
  }, [propertyId, user?.first_name, user?.last_name, fetchRoomGrid]);

  const filteredRooms = useMemo(() => {
    return rooms.filter((room: RoomBreakfastStatus) => {
      // Search filter
      if (searchQuery && !room.room_number.toLowerCase().includes(searchQuery.toLowerCase()) && 
          !room.guest_name?.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false;
      }

      // Status filter
      switch (filterStatus) {
        case 'has_breakfast':
          return room.has_guest && room.breakfast_package;
        case 'consumed':
          return room.consumed_today;
        case 'pending':
          return room.has_guest && room.breakfast_package && !room.consumed_today;
        case 'no_guest':
          return !room.has_guest;
        default:
          return true;
      }
    });
  }, [rooms, searchQuery, filterStatus]);

  const getRoomStatusColor = useCallback((room: RoomBreakfastStatus) => {
    if (!room.has_guest) return '#9E9E9E'; // Gray
    if (!room.breakfast_package) return '#FF9800'; // Orange
    if (room.consumed_today) return '#4CAF50'; // Green
    return '#2196F3'; // Blue
  }, []);

  const getRoomStatusText = useCallback((room: RoomBreakfastStatus) => {
    if (!room.has_guest) return 'No Guest';
    if (!room.breakfast_package) return 'No Breakfast Package';
    if (room.consumed_today) return 'Consumed';
    return 'Pending';
  }, []);

  const renderRoomCard = useCallback(({ item: room }: { item: RoomBreakfastStatus }) => (
    <Card style={[styles.roomCard, { borderLeftColor: getRoomStatusColor(room) }]}>
      <Card.Content>
        <View style={styles.roomHeader}>
          <Text variant="titleMedium" style={styles.roomNumber}>
            Room {room.room_number}
          </Text>
          <Badge style={{ backgroundColor: getRoomStatusColor(room) }}>
            {getRoomStatusText(room)}
          </Badge>
        </View>
        
        {room.has_guest && (
          <View style={styles.guestInfo}>
            <Text variant="bodyMedium" style={styles.guestName}>
              {room.guest_name}
            </Text>
            <Text variant="bodySmall" style={styles.roomDetails}>
              Floor {room.floor} • {room.room_type}
            </Text>
            {room.breakfast_package && (
              <Text variant="bodySmall" style={styles.breakfastInfo}>
                Breakfast Package: {room.breakfast_count} guests
              </Text>
            )}
            {room.consumed_today && room.consumed_at && (
              <Text variant="bodySmall" style={styles.consumedInfo}>
                Consumed at {new Date(room.consumed_at).toLocaleTimeString()}
                {room.consumed_by && ` by ${room.consumed_by}`}
              </Text>
            )}
          </View>
        )}
      </Card.Content>
      
      {room.has_guest && room.breakfast_package && !room.consumed_today && (
        <Card.Actions>
          <Button 
            mode="contained" 
            onPress={() => markBreakfastConsumed(room.room_number)}
            style={styles.consumeButton}
          >
            Mark Consumed
          </Button>
        </Card.Actions>
      )}
    </Card>
  ), [getRoomStatusColor, getRoomStatusText, markBreakfastConsumed]);

  if (loading && !refreshing) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" />
        <Text style={styles.loadingText}>Loading room grid...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text variant="headlineSmall" style={styles.title}>
          Room Grid Dashboard
        </Text>
        <Text variant="bodyMedium" style={styles.date}>
          {new Date(selectedDate).toLocaleDateString()}
        </Text>
      </View>

      <View style={styles.controls}>
        <Searchbar
          placeholder="Search rooms or guests..."
          onChangeText={setSearchQuery}
          value={searchQuery}
          style={styles.searchbar}
        />
        
        <View style={styles.filters}>
          <Menu
            visible={showFilterMenu}
            onDismiss={() => setShowFilterMenu(false)}
            anchor={
              <Chip 
                icon="filter" 
                onPress={() => setShowFilterMenu(true)}
                style={styles.filterChip}
              >
                {statusFilters.find(f => f.value === filterStatus)?.label}
              </Chip>
            }
          >
            {statusFilters.map(filter => (
              <Menu.Item
                key={filter.value}
                onPress={() => {
                  setFilterStatus(filter.value);
                  setShowFilterMenu(false);
                }}
                title={filter.label}
              />
            ))}
          </Menu>
        </View>
      </View>

      <View style={styles.stats}>
        <View style={styles.statItem}>
          <Text variant="labelLarge">{rooms.filter((r: RoomBreakfastStatus) => r.has_guest).length}</Text>
          <Text variant="bodySmall">Occupied</Text>
        </View>
        <View style={styles.statItem}>
          <Text variant="labelLarge">{rooms.filter((r: RoomBreakfastStatus) => r.breakfast_package).length}</Text>
          <Text variant="bodySmall">W/ Breakfast</Text>
        </View>
        <View style={styles.statItem}>
          <Text variant="labelLarge">{rooms.filter((r: RoomBreakfastStatus) => r.consumed_today).length}</Text>
          <Text variant="bodySmall">Consumed</Text>
        </View>
      </View>

      <FlatList
        data={filteredRooms}
        renderItem={renderRoomCard}
        keyExtractor={(item: RoomBreakfastStatus) => item.room_number}
        style={styles.roomList}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        showsVerticalScrollIndicator={false}
      />

      <FAB
        icon="sync"
        style={styles.fab}
        onPress={syncFromPMS}
        label="Sync PMS"
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
    backgroundColor: '#f5f5f5',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
  },
  header: {
    padding: 16,
    backgroundColor: '#ffffff',
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  title: {
    fontWeight: 'bold',
    color: '#FF6B35',
  },
  date: {
    marginTop: 4,
    color: '#666',
  },
  controls: {
    padding: 16,
    backgroundColor: '#ffffff',
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  searchbar: {
    marginBottom: 12,
  },
  filters: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  filterChip: {
    marginRight: 8,
  },
  stats: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    padding: 16,
    backgroundColor: '#ffffff',
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  statItem: {
    alignItems: 'center',
  },
  roomList: {
    flex: 1,
    padding: 16,
  },
  roomCard: {
    marginBottom: 12,
    borderLeftWidth: 4,
    elevation: 2,
  },
  roomHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  roomNumber: {
    fontWeight: 'bold',
    fontSize: 16,
  },
  guestInfo: {
    marginTop: 8,
  },
  guestName: {
    fontWeight: 'bold',
    marginBottom: 4,
  },
  roomDetails: {
    color: '#666',
    marginBottom: 2,
  },
  breakfastInfo: {
    color: '#2196F3',
    marginBottom: 2,
  },
  consumedInfo: {
    color: '#4CAF50',
    fontStyle: 'italic',
  },
  consumeButton: {
    marginLeft: 8,
  },
  fab: {
    position: 'absolute',
    margin: 16,
    right: 0,
    bottom: 0,
    backgroundColor: '#FF6B35',
  },
});
