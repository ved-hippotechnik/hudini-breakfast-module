// Room Grid Types for Property Management Dashboard

export interface Room {
  id: number;
  property_id: string;
  room_number: string;
  floor: number;
  room_type: string; // standard, deluxe, suite
  max_occupancy: number;
  status: string; // available, occupied, maintenance, out_of_order
  created_at: string;
  updated_at: string;
}

export interface Guest {
  id: number;
  pms_guest_id: string;
  reservation_id: string;
  room_number: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  check_in_date: string;
  check_out_date: string;
  adult_count: number;
  child_count: number;
  breakfast_package: boolean;
  breakfast_count: number;
  ohip_number?: string;
  property_id: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RoomBreakfastStatus {
  property_id: string;
  room_number: string;
  floor: number;
  room_type: string;
  status: string; // available, occupied, maintenance, out_of_order
  has_guest: boolean;
  guest_name: string;
  breakfast_package: boolean;
  breakfast_count: number;
  consumed_today: boolean;
  consumed_at?: string;
  consumed_by: string;
  check_in_date?: string;
  check_out_date?: string;
}

export interface DailyBreakfastConsumption {
  id: number;
  property_id: string;
  room_number: string;
  guest_id: number;
  consumption_date: string;
  consumed_at?: string;
  consumed_by?: number;
  status: string; // available, consumed, no_show
  notes: string;
  payment_method: string; // room_charge, ohip, comp, cash
  ohip_covered: boolean;
  pms_posted: boolean;
  pms_transaction_id: string;
  amount: number;
  created_at: string;
  updated_at: string;
}

export interface Staff {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  role: string; // staff, manager, admin
  property_id: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RoomGridResponse {
  rooms: RoomBreakfastStatus[];
  total_rooms: number;
  occupied_rooms: number;
  available_rooms: number;
  maintenance_rooms: number;
  rooms_with_breakfast: number;
  consumed_today: number;
  remaining_breakfasts: number;
}

export interface MarkConsumptionRequest {
  room_number: string;
  property_id: string;
  notes?: string;
  payment_method?: string;
}

export interface PMSSyncResponse {
  synced_guests: number;
  updated_rooms: number;
  errors: string[];
  message: string;
}

export interface ReportData {
  date: string;
  total_breakfasts: number;
  consumed_breakfasts: number;
  consumption_rate: number;
  ohip_transactions: number;
  revenue: number;
}

// Authentication types for staff
export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  staff: Staff;
  expires_at: string;
}

export interface AuthContextType {
  staff: Staff | null;
  token: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}
