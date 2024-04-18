
CREATE DOMAIN tracking_at AS timestamp with time zone NOT NULL DEFAULT now();
CREATE DOMAIN tracking_by AS text NOT NULL DEFAULT 'Unknown';