
CREATE DOMAIN text_medium AS varchar(256) NOT NULL;
CREATE DOMAIN text_medium_mandatory AS varchar(256) NOT NULL CHECK (value != '');

CREATE DOMAIN tracking_at AS timestamp with time zone NOT NULL DEFAULT now();
CREATE DOMAIN tracking_created_by AS text_medium_mandatory NOT NULL DEFAULT 'Unknown';
CREATE DOMAIN tracking_last_user_update_by AS text_medium NOT NULL DEFAULT '';