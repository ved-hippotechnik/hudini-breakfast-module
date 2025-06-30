import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Alert,
  TouchableOpacity,
} from 'react-native';
import { 
  Card, 
  Button, 
  TextInput, 
  Switch, 
  Divider,
  Avatar,
  List,
} from 'react-native-paper';
import { MaterialIcons } from '@expo/vector-icons';
import { useAuth } from '../../src/contexts/AuthContext';
import { theme } from '../../src/theme';
import { authAPI } from '../../src/services/api';

interface UserProfile {
  name: string;
  email: string;
  phone: string;
  dietaryRestrictions: string[];
  allergens: string[];
  preferences: {
    notifications: boolean;
    newsletter: boolean;
    locationServices: boolean;
    darkMode: boolean;
  };
  ohipInfo: {
    number: string;
    province: string;
    expiryDate: string;
  };
}

export default function ProfileScreen() {
  const { user, logout } = useAuth();
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  
  const [profile, setProfile] = useState<UserProfile>({
    name: user ? `${user.first_name} ${user.last_name}` : '',
    email: user?.email || '',
    phone: '',
    dietaryRestrictions: [],
    allergens: [],
    preferences: {
      notifications: true,
      newsletter: false,
      locationServices: true,
      darkMode: false,
    },
    ohipInfo: {
      number: '',
      province: 'Ontario',
      expiryDate: '',
    },
  });

  const dietaryOptions = [
    'Vegetarian', 'Vegan', 'Gluten-Free', 'Keto', 'Paleo', 'Low-Carb', 'Halal', 'Kosher'
  ];

  const allergenOptions = [
    'Nuts', 'Dairy', 'Eggs', 'Soy', 'Wheat', 'Shellfish', 'Fish', 'Sesame'
  ];

  const provinces = [
    'Alberta', 'British Columbia', 'Manitoba', 'New Brunswick', 'Newfoundland and Labrador',
    'Northwest Territories', 'Nova Scotia', 'Nunavut', 'Ontario', 'Prince Edward Island',
    'Quebec', 'Saskatchewan', 'Yukon'
  ];

  const handleSave = async () => {
    try {
      setLoading(true);
      // In a real implementation, this would call the API to update the profile
      // await apiService.updateProfile(profile);
      setEditing(false);
      Alert.alert('Success', 'Profile updated successfully');
    } catch (error) {
      Alert.alert('Error', 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    Alert.alert(
      'Logout',
      'Are you sure you want to logout?',
      [
        { text: 'Cancel', style: 'cancel' },
        { text: 'Logout', style: 'destructive', onPress: logout },
      ]
    );
  };

  const toggleDietaryRestriction = (restriction: string) => {
    setProfile(prev => ({
      ...prev,
      dietaryRestrictions: prev.dietaryRestrictions.includes(restriction)
        ? prev.dietaryRestrictions.filter(r => r !== restriction)
        : [...prev.dietaryRestrictions, restriction]
    }));
  };

  const toggleAllergen = (allergen: string) => {
    setProfile(prev => ({
      ...prev,
      allergens: prev.allergens.includes(allergen)
        ? prev.allergens.filter(a => a !== allergen)
        : [...prev.allergens, allergen]
    }));
  };

  const updatePreference = (key: keyof UserProfile['preferences'], value: boolean) => {
    setProfile(prev => ({
      ...prev,
      preferences: {
        ...prev.preferences,
        [key]: value
      }
    }));
  };

  return (
    <ScrollView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Avatar.Text 
          size={80} 
          label={profile.name.split(' ').map(n => n[0]).join('').toUpperCase() || 'U'} 
          style={styles.avatar}
        />
        <Text style={styles.name}>{profile.name || 'User'}</Text>
        <Text style={styles.email}>{profile.email}</Text>
        
        <View style={styles.headerButtons}>
          {editing ? (
            <>
              <Button 
                mode="contained" 
                onPress={handleSave} 
                loading={loading}
                style={styles.button}
              >
                Save
              </Button>
              <Button 
                mode="outlined" 
                onPress={() => setEditing(false)}
                style={styles.button}
              >
                Cancel
              </Button>
            </>
          ) : (
            <Button 
              mode="contained" 
              onPress={() => setEditing(true)}
              style={styles.button}
            >
              Edit Profile
            </Button>
          )}
        </View>
      </View>

      {/* Basic Information */}
      <Card style={styles.card}>
        <Card.Title title="Basic Information" />
        <Card.Content>
          <TextInput
            label="Full Name"
            value={profile.name}
            onChangeText={(text) => setProfile(prev => ({ ...prev, name: text }))}
            disabled={!editing}
            style={styles.input}
          />
          <TextInput
            label="Email"
            value={profile.email}
            onChangeText={(text) => setProfile(prev => ({ ...prev, email: text }))}
            disabled={!editing}
            keyboardType="email-address"
            style={styles.input}
          />
          <TextInput
            label="Phone Number"
            value={profile.phone}
            onChangeText={(text) => setProfile(prev => ({ ...prev, phone: text }))}
            disabled={!editing}
            keyboardType="phone-pad"
            style={styles.input}
          />
        </Card.Content>
      </Card>

      {/* OHIP Information */}
      <Card style={styles.card}>
        <Card.Title title="OHIP Information" />
        <Card.Content>
          <TextInput
            label="OHIP Number"
            value={profile.ohipInfo.number}
            onChangeText={(text) => setProfile(prev => ({ 
              ...prev, 
              ohipInfo: { ...prev.ohipInfo, number: text }
            }))}
            disabled={!editing}
            keyboardType="numeric"
            style={styles.input}
          />
          <TextInput
            label="Province"
            value={profile.ohipInfo.province}
            onChangeText={(text) => setProfile(prev => ({ 
              ...prev, 
              ohipInfo: { ...prev.ohipInfo, province: text }
            }))}
            disabled={!editing}
            style={styles.input}
          />
          <TextInput
            label="Expiry Date (YYYY-MM-DD)"
            value={profile.ohipInfo.expiryDate}
            onChangeText={(text) => setProfile(prev => ({ 
              ...prev, 
              ohipInfo: { ...prev.ohipInfo, expiryDate: text }
            }))}
            disabled={!editing}
            placeholder="2025-12-31"
            style={styles.input}
          />
        </Card.Content>
      </Card>

      {/* Dietary Restrictions */}
      <Card style={styles.card}>
        <Card.Title title="Dietary Restrictions" />
        <Card.Content>
          <View style={styles.tagContainer}>
            {dietaryOptions.map((option) => (
              <TouchableOpacity
                key={option}
                style={[
                  styles.tag,
                  profile.dietaryRestrictions.includes(option) && styles.tagSelected,
                  !editing && styles.tagDisabled
                ]}
                onPress={() => editing && toggleDietaryRestriction(option)}
                disabled={!editing}
              >
                <Text style={[
                  styles.tagText,
                  profile.dietaryRestrictions.includes(option) && styles.tagTextSelected
                ]}>
                  {option}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </Card.Content>
      </Card>

      {/* Allergens */}
      <Card style={styles.card}>
        <Card.Title title="Allergens" />
        <Card.Content>
          <View style={styles.tagContainer}>
            {allergenOptions.map((option) => (
              <TouchableOpacity
                key={option}
                style={[
                  styles.tag,
                  profile.allergens.includes(option) && styles.tagSelected,
                  !editing && styles.tagDisabled
                ]}
                onPress={() => editing && toggleAllergen(option)}
                disabled={!editing}
              >
                <Text style={[
                  styles.tagText,
                  profile.allergens.includes(option) && styles.tagTextSelected
                ]}>
                  {option}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </Card.Content>
      </Card>

      {/* Preferences */}
      <Card style={styles.card}>
        <Card.Title title="Preferences" />
        <Card.Content>
          <List.Item
            title="Push Notifications"
            description="Receive order updates and promotions"
            right={() => (
              <Switch
                value={profile.preferences.notifications}
                onValueChange={(value) => updatePreference('notifications', value)}
              />
            )}
          />
          <Divider />
          <List.Item
            title="Newsletter"
            description="Receive weekly nutrition tips and recipes"
            right={() => (
              <Switch
                value={profile.preferences.newsletter}
                onValueChange={(value) => updatePreference('newsletter', value)}
              />
            )}
          />
          <Divider />
          <List.Item
            title="Location Services"
            description="Find nearby breakfast locations"
            right={() => (
              <Switch
                value={profile.preferences.locationServices}
                onValueChange={(value) => updatePreference('locationServices', value)}
              />
            )}
          />
          <Divider />
          <List.Item
            title="Dark Mode"
            description="Use dark theme"
            right={() => (
              <Switch
                value={profile.preferences.darkMode}
                onValueChange={(value) => updatePreference('darkMode', value)}
              />
            )}
          />
        </Card.Content>
      </Card>

      {/* App Information */}
      <Card style={styles.card}>
        <Card.Title title="App Information" />
        <Card.Content>
          <List.Item
            title="Privacy Policy"
            left={() => <MaterialIcons name="privacy-tip" size={24} color="#666" />}
            right={() => <MaterialIcons name="chevron-right" size={24} color="#666" />}
            onPress={() => {/* Navigate to privacy policy */}}
          />
          <Divider />
          <List.Item
            title="Terms of Service"
            left={() => <MaterialIcons name="description" size={24} color="#666" />}
            right={() => <MaterialIcons name="chevron-right" size={24} color="#666" />}
            onPress={() => {/* Navigate to terms */}}
          />
          <Divider />
          <List.Item
            title="Help & Support"
            left={() => <MaterialIcons name="help" size={24} color="#666" />}
            right={() => <MaterialIcons name="chevron-right" size={24} color="#666" />}
            onPress={() => {/* Navigate to help */}}
          />
          <Divider />
          <List.Item
            title="App Version"
            description="1.0.0"
            left={() => <MaterialIcons name="info" size={24} color="#666" />}
          />
        </Card.Content>
      </Card>

      {/* Logout Button */}
      <Card style={styles.card}>
        <Card.Content>
          <Button
            mode="outlined"
            onPress={handleLogout}
            style={[styles.button, styles.logoutButton]}
            labelStyle={styles.logoutButtonText}
            icon="logout"
          >
            Logout
          </Button>
        </Card.Content>
      </Card>

      <View style={styles.bottomSpacing} />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  header: {
    alignItems: 'center',
    padding: 20,
    backgroundColor: '#fff',
    marginBottom: 16,
  },
  avatar: {
    backgroundColor: theme.colors.primary,
    marginBottom: 16,
  },
  name: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 4,
  },
  email: {
    fontSize: 16,
    color: '#666',
    marginBottom: 20,
  },
  headerButtons: {
    flexDirection: 'row',
    gap: 12,
  },
  button: {
    marginVertical: 8,
  },
  card: {
    margin: 16,
    backgroundColor: '#fff',
  },
  input: {
    marginBottom: 12,
    backgroundColor: '#fff',
  },
  tagContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  tag: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: '#ddd',
    backgroundColor: '#f8f9fa',
  },
  tagSelected: {
    backgroundColor: theme.colors.primary,
    borderColor: theme.colors.primary,
  },
  tagDisabled: {
    opacity: 0.6,
  },
  tagText: {
    fontSize: 12,
    color: '#666',
  },
  tagTextSelected: {
    color: '#fff',
  },
  logoutButton: {
    borderColor: '#dc3545',
  },
  logoutButtonText: {
    color: '#dc3545',
  },
  bottomSpacing: {
    height: 20,
  },
});
