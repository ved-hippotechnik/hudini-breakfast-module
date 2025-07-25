import React, { useState, useEffect } from 'react';
import {
  View,
  StyleSheet,
  ScrollView,
  Alert,
  RefreshControl,
  Dimensions,
  TouchableOpacity,
} from 'react-native';
import {
  Text,
  Button,
  Badge,
  ActivityIndicator,
  Searchbar,
  Chip,
  Portal,
  Modal,
  Surface,
  Divider,
} from 'react-native-paper';
import { roomGridAPI } from '../../src/services/api';
import { useAuth } from '../../src/contexts/AuthContext';
import { RoomBreakfastStatus } from '../../src/types/roomgrid';

const { width } = Dimensions.get('window');
const ROOM_CARD_SIZE = (width - 60) / 4; // 4 rooms per row with margins

export default function RoomGridDashboard() {
  const [rooms, setRooms] = useState<RoomBreakfastStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [selectedRoom, setSelectedRoom] = useState<RoomBreakfastStatus | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [selectedDate, setSelectedDate] = useState(new Date().toISOString().split('T')[0]);
  
  const { user } = useAuth();
  const propertyId = user?.property_id || 'PROP001';

  const statusFilters = [
    { value: 'all', label: 'All', color: '#6c757d' },
    { value: 'vacant', label: 'Vacant', color: '#28a745' },
    { value: 'occupied', label: 'Occupied', color: '#ffc107' },
    { value: 'breakfast', label: 'Breakfast', color: '#007bff' },
    { value: 'consumed', label: 'Consumed', color: '#17a2b8' },
    { value: 'maintenance', label: 'Maintenance', color: '#dc3545' },
  ];

  useEffect(() => {
    fetchRoomGrid();
  }, [selectedDate]);

  const fetchRoomGrid = async () => {
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
  };

  const onRefresh = async () => {
    setRefreshing(true);
    await fetchRoomGrid();
    setRefreshing(false);
  };

  const syncFromPMS = async () => {
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
  };

  const markBreakfastConsumed = async (roomNumber: string) => {
    try {
      const data = {
        property_id: propertyId,
        room_number: roomNumber,
        payment_method: 'room_charge',
        notes: `Consumed via mobile app by ${user?.first_name} ${user?.last_name}`,
      };
      
      await roomGridAPI.markBreakfastConsumed(data);
      await fetchRoomGrid();
      setModalVisible(false);
      Alert.alert('Success', `Breakfast marked as consumed for room ${roomNumber}`);
    } catch (error: any) {
      console.error('Error marking breakfast consumed:', error);
      const message = error.response?.data?.error || 'Failed to mark breakfast as consumed';
      Alert.alert('Error', message);
    }
  };

  const getRoomStatusColor = (room: RoomBreakfastStatus) => {
    if (room.status === 'maintenance') return '#dc3545';
    if (room.status === 'out_of_order') return '#6c757d';
    if (!room.has_guest) return '#28a745';
    if (room.consumed_today) return '#17a2b8';
    if (room.breakfast_package) return '#007bff';
    return '#ffc107';
  };

  const getRoomStatusText = (room: RoomBreakfastStatus) => {
    if (room.status === 'maintenance') return 'Maintenance';
    if (room.status === 'out_of_order') return 'Out of Order';
    if (!room.has_guest) return 'Vacant';
    if (room.consumed_today) return 'Consumed';
    if (room.breakfast_package) return 'Breakfast';
    return 'Occupied';
  };

  const filteredRooms = rooms.filter((room: RoomBreakfastStatus) => {
    const matchesSearch = room.room_number.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         (room.guest_name && room.guest_name.toLowerCase().includes(searchQuery.toLowerCase()));
    
    const matchesFilter = (() => {
      switch (filterStatus) {
        case 'vacant': return !room.has_guest;
        case 'occupied': return room.has_guest && !room.breakfast_package;
        case 'breakfast': return room.breakfast_package && !room.consumed_today;
        case 'consumed': return room.consumed_today;
        case 'maintenance': return room.status === 'maintenance' || room.status === 'out_of_order';
        default: return true;
      }
    })();

    return matchesSearch && matchesFilter;
  });

  // Group rooms by floor
  const roomsByFloor = filteredRooms.reduce((acc, room) => {
    const floor = room.floor || 1;
    if (!acc[floor]) acc[floor] = [];
    acc[floor].push(room);
    return acc;
  }, {} as Record<number, RoomBreakfastStatus[]>);

  const sortedFloors = Object.keys(roomsByFloor)
    .map(Number)
    .sort((a, b) => b - a); // Highest floor first

  const getStats = () => {
    const total = rooms.length;
    const occupied = rooms.filter(r => r.has_guest).length;
    const breakfast = rooms.filter(r => r.breakfast_package).length;
    const consumed = rooms.filter(r => r.consumed_today).length;
    const vacant = rooms.filter(r => !r.has_guest).length;
    const pending = rooms.filter(r => r.breakfast_package && !r.consumed_today).length;

    return { total, occupied, breakfast, consumed, vacant, pending };
  };

  const stats = getStats();

  const openRoomModal = (room: RoomBreakfastStatus) => {
    setSelectedRoom(room);
    setModalVisible(true);
  };

  if (loading && rooms.length === 0) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#007bff" />
        <Text style={styles.loadingText}>Loading room grid...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {/* Header with Stats */}
      <Surface style={styles.statsContainer}>
        <Text style={styles.dashboardTitle}>🏨 Room Grid Dashboard</Text>
        <View style={styles.statsGrid}>
          <View style={styles.statItem}>
            <Text style={styles.statNumber}>{stats.total}</Text>
            <Text style={styles.statLabel}>Total</Text>
          </View>
          <View style={styles.statItem}>
            <Text style={[styles.statNumber, { color: '#ffc107' }]}>{stats.occupied}</Text>
            <Text style={styles.statLabel}>Occupied</Text>
          </View>
          <View style={styles.statItem}>
            <Text style={[styles.statNumber, { color: '#007bff' }]}>{stats.breakfast}</Text>
            <Text style={styles.statLabel}>Breakfast</Text>
          </View>
          <View style={styles.statItem}>
            <Text style={[styles.statNumber, { color: '#17a2b8' }]}>{stats.consumed}</Text>
            <Text style={styles.statLabel}>Consumed</Text>
          </View>
        </View>
      </Surface>

      {/* Search and Filters */}
      <View style={styles.controlsContainer}>
        <Searchbar
          placeholder="Search rooms or guests..."
          onChangeText={setSearchQuery}
          value={searchQuery}
          style={styles.searchbar}
        />
        
        <ScrollView 
          horizontal 
          showsHorizontalScrollIndicator={false}
          style={styles.filtersContainer}
        >
          {statusFilters.map((filter) => (
            <Chip
              key={filter.value}
              mode={filterStatus === filter.value ? 'flat' : 'outlined'}
              selected={filterStatus === filter.value}
              onPress={() => setFilterStatus(filter.value)}
              style={[
                styles.filterChip,
                filterStatus === filter.value && { backgroundColor: filter.color }
              ]}
              textStyle={filterStatus === filter.value ? { color: 'white' } : {}}
            >
              {filter.label}
            </Chip>
          ))}
        </ScrollView>

        <Button
          mode="contained"
          onPress={syncFromPMS}
          style={styles.syncButton}
          icon="sync"
        >
          Sync PMS
        </Button>
      </View>

      {/* Room Grid */}
      <ScrollView
        style={styles.gridContainer}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
      >
        {sortedFloors.map((floor) => (
          <View key={floor} style={styles.floorSection}>
            <Surface style={styles.floorHeader}>
              <Text style={styles.floorTitle}>Floor {floor}</Text>
              <Badge>{roomsByFloor[floor].length} rooms</Badge>
            </Surface>
            
            <View style={styles.roomsGrid}>
              {roomsByFloor[floor]
                .sort((a, b) => a.room_number.localeCompare(b.room_number, undefined, { numeric: true }))
                .map((room) => (
                  <TouchableOpacity
                    key={room.room_number}
                    style={[
                      styles.roomCard,
                      { 
                        borderColor: getRoomStatusColor(room),
                        backgroundColor: `${getRoomStatusColor(room)}15` // 15% opacity
                      }
                    ]}
                    onPress={() => openRoomModal(room)}
                  >
                    <View style={styles.roomHeader}>
                      <Text style={styles.roomNumber}>{room.room_number}</Text>
                      {room.breakfast_package && (
                        <View style={[
                          styles.breakfastIndicator,
                          { backgroundColor: room.consumed_today ? '#17a2b8' : '#007bff' }
                        ]} />
                      )}
                    </View>
                    
                    <Text style={[
                      styles.roomStatus,
                      { color: getRoomStatusColor(room) }
                    ]}>
                      {getRoomStatusText(room)}
                    </Text>
                    
                    {room.has_guest && (
                      <Text style={styles.guestName} numberOfLines={1}>
                        {room.guest_name}
                      </Text>
                    )}
                  </TouchableOpacity>
                ))}
            </View>
          </View>
        ))}
      </ScrollView>

      {/* Room Details Modal */}
      <Portal>
        <Modal
          visible={modalVisible}
          onDismiss={() => setModalVisible(false)}
          contentContainerStyle={styles.modalContainer}
        >
          {selectedRoom && (
            <ScrollView>
              <Text style={styles.modalTitle}>Room {selectedRoom.room_number}</Text>
              
              <View style={styles.modalSection}>
                <Text style={styles.modalLabel}>Floor:</Text>
                <Text style={styles.modalValue}>{selectedRoom.floor}</Text>
              </View>

              <View style={styles.modalSection}>
                <Text style={styles.modalLabel}>Room Type:</Text>
                <Text style={styles.modalValue}>{selectedRoom.room_type || 'Standard'}</Text>
              </View>

              <View style={styles.modalSection}>
                <Text style={styles.modalLabel}>Status:</Text>
                <Text style={[styles.modalValue, { color: getRoomStatusColor(selectedRoom) }]}>
                  {getRoomStatusText(selectedRoom)}
                </Text>
              </View>

              {selectedRoom.has_guest && (
                <>
                  <Divider style={styles.modalDivider} />
                  
                  <View style={styles.modalSection}>
                    <Text style={styles.modalLabel}>Guest Name:</Text>
                    <Text style={styles.modalValue}>{selectedRoom.guest_name}</Text>
                  </View>

                  <View style={styles.modalSection}>
                    <Text style={styles.modalLabel}>Check-in:</Text>
                    <Text style={styles.modalValue}>
                      {selectedRoom.check_in_date ? new Date(selectedRoom.check_in_date).toLocaleDateString() : 'N/A'}
                    </Text>
                  </View>

                  <View style={styles.modalSection}>
                    <Text style={styles.modalLabel}>Check-out:</Text>
                    <Text style={styles.modalValue}>
                      {selectedRoom.check_out_date ? new Date(selectedRoom.check_out_date).toLocaleDateString() : 'N/A'}
                    </Text>
                  </View>

                  <View style={styles.modalSection}>
                    <Text style={styles.modalLabel}>Breakfast Package:</Text>
                    <Text style={[styles.modalValue, { color: selectedRoom.breakfast_package ? '#007bff' : '#6c757d' }]}>
                      {selectedRoom.breakfast_package ? 'Yes' : 'No'}
                    </Text>
                  </View>

                  {selectedRoom.breakfast_package && (
                    <>
                      <View style={styles.modalSection}>
                        <Text style={styles.modalLabel}>Breakfast Count:</Text>
                        <Text style={styles.modalValue}>{selectedRoom.breakfast_count || 0}</Text>
                      </View>

                      <View style={styles.modalSection}>
                        <Text style={styles.modalLabel}>Consumed Today:</Text>
                        <Text style={[styles.modalValue, { color: selectedRoom.consumed_today ? '#17a2b8' : '#dc3545' }]}>
                          {selectedRoom.consumed_today ? 'Yes' : 'No'}
                        </Text>
                      </View>

                      {selectedRoom.consumed_today && selectedRoom.consumed_at && (
                        <>
                          <View style={styles.modalSection}>
                            <Text style={styles.modalLabel}>Consumed At:</Text>
                            <Text style={styles.modalValue}>
                              {new Date(selectedRoom.consumed_at).toLocaleString()}
                            </Text>
                          </View>

                          <View style={styles.modalSection}>
                            <Text style={styles.modalLabel}>Consumed By:</Text>
                            <Text style={styles.modalValue}>{selectedRoom.consumed_by || 'N/A'}</Text>
                          </View>
                        </>
                      )}
                    </>
                  )}
                </>
              )}

              <View style={styles.modalActions}>
                {selectedRoom.breakfast_package && !selectedRoom.consumed_today && (
                  <Button
                    mode="contained"
                    onPress={() => markBreakfastConsumed(selectedRoom.room_number)}
                    style={[styles.modalButton, { backgroundColor: '#28a745' }]}
                  >
                    Mark as Consumed
                  </Button>
                )}
                
                <Button
                  mode="outlined"
                  onPress={() => setModalVisible(false)}
                  style={styles.modalButton}
                >
                  Close
                </Button>
              </View>
            </ScrollView>
          )}
        </Modal>
      </Portal>
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
    color: '#666',
  },
  statsContainer: {
    margin: 16,
    padding: 16,
    borderRadius: 12,
    elevation: 2,
  },
  dashboardTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    textAlign: 'center',
    marginBottom: 16,
    color: '#333',
  },
  statsGrid: {
    flexDirection: 'row',
    justifyContent: 'space-around',
  },
  statItem: {
    alignItems: 'center',
  },
  statNumber: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333',
  },
  statLabel: {
    fontSize: 12,
    color: '#666',
    marginTop: 4,
  },
  controlsContainer: {
    paddingHorizontal: 16,
    paddingBottom: 16,
  },
  searchbar: {
    marginBottom: 12,
    elevation: 1,
  },
  filtersContainer: {
    marginBottom: 12,
  },
  filterChip: {
    marginRight: 8,
  },
  syncButton: {
    backgroundColor: '#007bff',
  },
  gridContainer: {
    flex: 1,
  },
  floorSection: {
    marginBottom: 20,
  },
  floorHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 12,
    marginHorizontal: 16,
    marginBottom: 8,
    borderRadius: 8,
    elevation: 1,
  },
  floorTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
  },
  roomsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    paddingHorizontal: 12,
  },
  roomCard: {
    width: ROOM_CARD_SIZE,
    height: ROOM_CARD_SIZE * 0.8,
    backgroundColor: 'white',
    borderRadius: 8,
    borderWidth: 2,
    margin: 4,
    padding: 8,
    elevation: 2,
    justifyContent: 'space-between',
  },
  roomHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
  },
  roomNumber: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
  },
  breakfastIndicator: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  roomStatus: {
    fontSize: 10,
    fontWeight: '600',
    textTransform: 'uppercase',
    textAlign: 'center',
  },
  guestName: {
    fontSize: 10,
    color: '#666',
    textAlign: 'center',
    marginTop: 4,
  },
  modalContainer: {
    backgroundColor: 'white',
    margin: 20,
    padding: 20,
    borderRadius: 12,
    maxHeight: '80%',
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 16,
    textAlign: 'center',
  },
  modalSection: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
  },
  modalLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: '#666',
    flex: 1,
  },
  modalValue: {
    fontSize: 14,
    color: '#333',
    flex: 1,
    textAlign: 'right',
  },
  modalDivider: {
    marginVertical: 12,
  },
  modalActions: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    marginTop: 20,
    gap: 12,
  },
  modalButton: {
    flex: 1,
  },
});
