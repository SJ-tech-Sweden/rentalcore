-- Enhanced Equipment Packages Migration (MySQL Compatible)
-- Add new fields for production-ready equipment packages

-- Add new columns to equipment_packages table
-- Note: Run each ALTER TABLE separately if you get syntax errors
-- Add max_rental_days column
-- Add category column  
-- Add tags column
-- Add last_used_at column
-- Add total_revenue column
-- Add indexes for better performance
-- Add indexes for package_devices table
-- Update existing package_items to be valid JSON if NULL
-- Sample equipment packages removed for production use
-- Create packages through the application interface as needed
DO $$
BEGIN
	RAISE NOTICE 'Skipping MySQL-original migration 007_equipment_packages.sql; converted migration should be used instead.';
END$$;