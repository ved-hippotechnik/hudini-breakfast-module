// API Response Types
export interface APIResponse<T = any> {
  success: boolean;
  message?: string;
  data?: T;
  error?: APIError;
  timestamp: string;
  request_id?: string;
}

export interface APIError {
  code: string;
  message: string;
  details?: string;
}

export interface PaginatedResponse<T = any> extends APIResponse<T> {
  pagination?: PaginationMeta;
}

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

// Authentication Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
  expires_at: string;
}

export interface User {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  property_id: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Guest Types
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

export interface CreateGuestRequest {
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
}

export interface UpdateGuestRequest extends Partial<CreateGuestRequest> {
  id: number;
}

// Room Types
export interface Room {
  id: number;
  property_id: string;
  room_number: string;
  floor: number;
  room_type: string;
  max_occupancy: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface RoomBreakfastStatus {
  property_id: string;
  room_number: string;
  floor: number;
  room_type: string;
  status: string;
  has_guest: boolean;
  guest_name: string;
  breakfast_package: boolean;
  breakfast_count: number;
  consumed_today: boolean;
  consumed_at?: string;
  consumed_by?: string;
  check_in_date?: string;
  check_out_date?: string;
}

// Breakfast Consumption Types
export interface DailyBreakfastConsumption {
  id: number;
  property_id: string;
  room_number: string;
  guest_id: number;
  consumption_date: string;
  consumed_at?: string;
  consumed_by?: number;
  status: string;
  notes?: string;
  payment_method: string;
  ohip_covered: boolean;
  pms_posted: boolean;
  pms_transaction_id?: string;
  amount: number;
  created_at: string;
  updated_at: string;
  guest?: Guest;
  staff?: User;
  room?: Room;
}

export interface MarkConsumptionRequest {
  property_id: string;
  room_number: string;
  payment_method: string;
  notes?: string;
}

// Analytics Types
export interface DailyBreakfastReport {
  date: string;
  total_rooms_with_breakfast: number;
  total_consumed: number;
  total_not_consumed: number;
  consumption_rate: number;
  ohip_covered_count: number;
  pms_charges_posted: number;
}

export interface BreakfastAnalytics {
  period: string;
  daily_trend: DailyTrend[];
}

export interface DailyTrend {
  date: string;
  consumed: number;
  total: number;
  rate: number;
}

export interface AdvancedAnalytics {
  total_revenue: number;
  avg_consumption_rate: number;
  peak_consumption_time: string;
  top_consuming_rooms: string[];
  forecasted_demand: number;
  waste_percentage: number;
  cost_per_serving: number;
  monthly_trends: MonthlyTrend[];
}

export interface MonthlyTrend {
  month: string;
  consumption_rate: number;
  revenue: number;
  waste_percentage: number;
}

export interface RealtimeMetrics {
  current_consumption_rate: number;
  rooms_consumed_today: number;
  rooms_pending: number;
  estimated_completion_time: string;
  live_updates: LiveUpdate[];
}

export interface LiveUpdate {
  timestamp: string;
  room_number: string;
  action: string;
  staff_member: string;
}

// PMS Integration Types
export interface PMSIntegrationStatus {
  status: string;
  last_sync: string;
  total_synced: number;
  errors: string[];
}

export interface SyncFromPMSRequest {
  property_id: string;
  force_sync?: boolean;
}

// WebSocket Types
export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
}

export interface RoomUpdateMessage {
  type: 'room_update';
  data: {
    room_number: string;
    property_id: string;
    status: RoomBreakfastStatus;
  };
}

export interface ConsumptionUpdateMessage {
  type: 'consumption_update';
  data: {
    room_number: string;
    property_id: string;
    consumed_at: string;
    consumed_by: string;
  };
}

// Filter and Search Types
export interface RoomFilter {
  status: 'all' | 'has_breakfast' | 'consumed' | 'pending' | 'no_guest';
  floor?: number;
  room_type?: string;
  search_query?: string;
}

export interface DateRange {
  start_date: string;
  end_date: string;
}

export interface AnalyticsFilter {
  property_id: string;
  period: 'today' | 'week' | 'month' | 'custom';
  date_range?: DateRange;
}

// Error Types
export interface ValidationError {
  field: string;
  message: string;
  code: string;
}

export interface APIErrorResponse {
  success: false;
  error: APIError;
  timestamp: string;
  request_id?: string;
}

// Utility Types
export type LoadingState = 'idle' | 'loading' | 'success' | 'error';

export interface AsyncState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

export interface PaginationState {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

// Request/Response wrapper types
export type APIRequest<T = any> = T;
export type APIResponseData<T = any> = APIResponse<T>;
export type PaginatedAPIResponse<T = any> = PaginatedResponse<T>;