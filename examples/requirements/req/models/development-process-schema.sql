/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `family`;

CREATE TABLE `family` ( # core partitioning of the database
  `family_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL default '',
  PRIMARY KEY (`family_id`),
  UNIQUE KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `phase`;

CREATE TABLE `phase` ( # fundamental phase skeleton for process family
  `phase_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `family_id` int(11) unsigned NOT NULL,
  `num` int(11) unsigned NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL default '',
  PRIMARY KEY (`phase_id`),
  UNIQUE KEY (`family_id`,`num`),
  UNIQUE KEY (`family_id`,`name`),
  CONSTRAINT `fk_phase_family` FOREIGN KEY (`family_id`) REFERENCES `family` (`family_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `defect_type`;

CREATE TABLE `defect_type` ( # type of defect
  `defect_type_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `family_id` int(11) unsigned NOT NULL,
  `num` int(11) unsigned NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  `base_num` int(11) unsigned NOT NULL,
  PRIMARY KEY (`defect_type_id`),
  UNIQUE KEY (`family_id`, `num`),
  UNIQUE KEY (`family_id`, `name`),
  CONSTRAINT `fk_defect_type_family` FOREIGN KEY (`family_id`) REFERENCES `family` (`family_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `language`;

CREATE TABLE `language` ( # programming language
  `language_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `family_id` int(11) unsigned NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  PRIMARY KEY (`language_id`),
  UNIQUE KEY (`family_id`, `name`),
  INDEX (`language_id`,`family_id`),
  CONSTRAINT `fk_language_family` FOREIGN KEY (`family_id`) REFERENCES `family` (`family_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `process`;

CREATE TABLE `process` ( # a process to follow
  `process_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `family_id` int(11) unsigned NOT NULL,
  `name` varchar(255) NOT NULL,
  `version` int(11) unsigned NOT NULL,
  `version_minor` int(11) unsigned NOT NULL DEFAULT '0',
  `purpose` varchar(255) NOT NULL,
  `entry_criteria` TEXT NOT NULL,
  `exit_criteria` TEXT NOT NULL,
  `script_lock` tinyint(1) unsigned NOT NULL DEFAULT FALSE,
  `ancestor_id` int(11) unsigned,
  PRIMARY KEY (`process_id`),
  UNIQUE KEY (`family_id`, `name`, `version`, `version_minor`),
  CONSTRAINT `fk_process_family` FOREIGN KEY (`family_id`) REFERENCES `family` (`family_id`),
  CONSTRAINT `fk_process_ancestor` FOREIGN KEY (`ancestor_id`) REFERENCES `process` (`process_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `script`;

CREATE TABLE `script` ( # a step-by-step process script
  `script_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `process_id` int(11) unsigned NOT NULL,
  `num` int(11) unsigned NOT NULL,
  `name` varchar(255) NOT NULL,
  `task_summary` TEXT NOT NULL,
  `purpose` varchar(255) NOT NULL,
  `entry_criteria` TEXT NOT NULL,
  `exit_criteria` TEXT NOT NULL,
  `cycle` tinyint(1) unsigned NOT NULL DEFAULT FALSE,
  PRIMARY KEY (`script_id`),
  UNIQUE KEY (`process_id`, `num`),
  UNIQUE KEY (`process_id`, `name`),
  CONSTRAINT `fk_script_process` FOREIGN KEY (`process_id`) REFERENCES `process` (`process_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `step`;

CREATE TABLE `step` ( # a step-by-step process script
  `step_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `script_id` int(11) unsigned NOT NULL,
  `num` int(11) unsigned NOT NULL,
  `name` varchar(255) NOT NULL,
  `phase_id` int(11) unsigned NOT NULL,
  `tasks` TEXT NOT NULL,
  PRIMARY KEY (`step_id`),
  UNIQUE KEY (`script_id`, `num`),
  UNIQUE KEY (`script_id`, `name`),
  CONSTRAINT `fk_step_script` FOREIGN KEY (`script_id`) REFERENCES `script` (`script_id`),
  CONSTRAINT `fk_step_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `estimate`;

CREATE TABLE `estimate` ( # an estimate of size or time
  `estimate_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `family_id` int(11) unsigned NOT NULL,
  # How the estimate is categorized.
  `language_id` int(11) unsigned NOT NULL,
  `axis` enum('size','time') NOT NULL,
  `scope` enum('project','phase','cycle','proxy','added','modified','deleted') NOT NULL,
  # The version of this estimate.
  `version` int(11) unsigned NOT NULL DEFAULT '1', # Always increment as we update. 
  # The estimate itself.
  `mean` float(13,2) NOT NULL, # The middle of the estimate bell curve.
  `variance` float(13,2) unsigned NOT NULL, # The shape of the estimate bell curve.
  `low` float(13,2) NOT NULL,  # The lowest estimate based on the prediction interval.
  `high` float(13,2) NOT NULL, # The highest estimate based on the prediction interval.
  # Eventually we'll know exactly how much work was needed.
  `actual` int(11) unsigned NOT NULL DEFAULT '0', # The value recorded when work completed.
  # Values common to all estimation methods.
  `method` enum('guess','sum','portion') NOT NULL, # 'average','fuzzy','linear'
  `prediction_interval` decimal(5,4) unsigned NOT NULL, # What high, low range are we interested in: 0.0000 to 1.0000
  `comment` varchar(255) NOT NULL, # Any note about how the estimate was computed.
  `estimation_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, # The time the estimate was made.
  # Values used only by a pure guess estimate.
  `guess_lowest_80pred` int(11) unsigned, # The lowest we expect it to be at 80% prediction interval.
  `guess_highest_80pred` int(11) unsigned, # The highest we expect it to be at 80% prediction interval.
  # Values used only by a sum estimate, summing other estimates each with their own variance.
  `sum_count` int(11) unsigned , # The number of inputs into the computation.
  `sum_mean` decimal(13,2), # The middle of the estimate bell curve.
  `sum_variance` decimal(13,2) unsigned, # The shape of the estimate bell curve.
  # Values used only by a portion estimate, some fraction of another estimate.
  `portion_portion` decimal(5,4) unsigned, # The fraction of the original value: 0.0000 to 1.0000
  `portion_mean` decimal(13,2), # The middle of the estimate bell curve.
  `portion_variance` decimal(13,2) unsigned, # The shape of the estimate bell curve.
  # Values used only by a average estimate, averaging historic values.
  #`average_count` int(11) unsigned , # The number of inputs into the computation.
  #`average_mean` decimal(13,2) unsigned, # The middle of the estimate bell curve.
  #`average_low_1dev` decimal(13,2) unsigned,  # The lowest estimate based on the prediction interval.
  #`average_high_1dev` decimal(13,2) unsigned, # The highest estimate based on the prediction interval.
  # Values used only be a fuzzy category estimate.
  #`fuzzy_category` enum('very_small', 'small', 'medium', 'large', 'very_large'), # The category.
  # Values used only by linear regression estimates.
  #`linear_b0` decimal(13,2) unsigned, # Where the regression line crosses the axis.
  #`linear_b1` decimal(13,2) unsigned, # The angle of the regression line.
  PRIMARY KEY (`estimate_id`),
  INDEX (`actual`),
  CONSTRAINT `fk_estimate_family` FOREIGN KEY (`family_id`) REFERENCES `family` (`family_id`),
  CONSTRAINT `fk_estimate_language` FOREIGN KEY (`language_id`,`family_id`) REFERENCES `language` (`language_id`,`family_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

# ------------------------------------------------------------

DROP TABLE IF EXISTS `estimate_historic`;

CREATE TABLE `estimate_historic` ( # a history of past estimate iterations
  `estimate_id` int(11) unsigned NOT NULL,
  `family_id` int(11) unsigned NOT NULL,
  # How the estimate is categorized.
  `language_id` int(11) unsigned NOT NULL,
  `axis` enum('size','time') NOT NULL,
  `scope` enum('project','phase','cycle','proxy','added','modified','deleted') NOT NULL,
  # The version of this estimate.
  `version` int(11) unsigned NOT NULL DEFAULT '1', # Always increment as we update. 
  # The estimate itself.
  `mean` float(13,2) NOT NULL, # The middle of the estimate bell curve.
  `variance` float(13,2) unsigned NOT NULL, # The shape of the estimate bell curve.
  `low` float(13,2) NOT NULL,  # The lowest estimate based on the prediction interval.
  `high` float(13,2) NOT NULL, # The highest estimate based on the prediction interval.
  # Eventually we'll know exactly how much work was needed.
  `actual` int(11) unsigned NOT NULL DEFAULT '0', # The value recorded when work completed.
  # Values common to all estimation methods.
  `method` enum('guess','sum','portion') NOT NULL, # 'average','fuzzy','linear'
  `prediction_interval` decimal(5,4) unsigned NOT NULL, # What high, low range are we interested in: 0.0000 to 1.0000
  `comment` varchar(255) NOT NULL, # Any note about how the estimate was computed.
  `estimation_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, # The time the estimate was made.
  # Values used only by a pure guess estimate.
  `guess_lowest_80pred` int(11) unsigned, # The lowest we expect it to be at 80% prediction interval.
  `guess_highest_80pred` int(11) unsigned, # The highest we expect it to be at 80% prediction interval.
  # Values used only by a sum estimate, summing other estimates each with their own variance.
  `sum_count` int(11) unsigned , # The number of inputs into the computation.
  `sum_mean` decimal(13,2), # The middle of the estimate bell curve.
  `sum_variance` decimal(13,2) unsigned, # The shape of the estimate bell curve.
  # Values used only by a portion estimate, some fraction of another estimate.
  `portion_portion` decimal(5,4) unsigned, # The fraction of the original value: 0.0000 to 1.0000
  `portion_mean` decimal(13,2), # The middle of the estimate bell curve.
  `portion_variance` decimal(13,2) unsigned, # The shape of the estimate bell curve.
  # Values used only by a average estimate, averaging historic values.
  #`average_count` int(11) unsigned , # The number of inputs into the computation.
  #`average_mean` decimal(13,2) unsigned, # The middle of the estimate bell curve.
  #`average_low_1dev` decimal(13,2) unsigned,  # The lowest estimate based on the prediction interval.
  #`average_high_1dev` decimal(13,2) unsigned, # The highest estimate based on the prediction interval.
  # Values used only be a fuzzy category estimate.
  #`fuzzy_category` enum('very_small', 'small', 'medium', 'large', 'very_large'), # The category.
  # Values used only by linear regression estimates.
  #`linear_b0` decimal(13,2) unsigned, # Where the regression line crosses the axis.
  #`linear_b1` decimal(13,2) unsigned, # The angle of the regression line.
  PRIMARY KEY (`estimate_id`,`version`),
  CONSTRAINT `fk_estimate_historic_estimate` FOREIGN KEY (`estimate_id`) REFERENCES `estimate` (`estimate_id`),
  CONSTRAINT `fk_estimate_historic_family` FOREIGN KEY (`family_id`) REFERENCES `family` (`family_id`),
  CONSTRAINT `fk_estimate_historic_language` FOREIGN KEY (`language_id`,`family_id`) REFERENCES `language` (`language_id`,`family_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `method`;
-- 
-- CREATE TABLE `method` ( # Programming method or language.
--   `method_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL default '',
--   PRIMARY KEY (`method_id`),
--   UNIQUE KEY (`name`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 


-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `project`;
-- 
-- CREATE TABLE `project` (
--   `project_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
--   `name` varchar(32) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   `process_id` int(11) unsigned NOT NULL,
--   `started_time` DATETIME, # If null, not started.
--   `estimate_minute` int(11) unsigned NOT NULL DEFAULT '0',
--   `estimate_comment` varchar(255) NOT NULL DEFAULT '',
--   PRIMARY KEY (`project_id`),
--   CONSTRAINT `fk_project_process` FOREIGN KEY (`process_id`) REFERENCES `process` (`process_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `project_stat_phase`;
-- 
-- CREATE TABLE `project_stat_phase` (
--   `project_stat_phase_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `stat_phase_id` int(11) unsigned NOT NULL,
--   `bucket_id` int(11) unsigned NOT NULL,
--   `method_id` int(11) unsigned,
--   `estimate_minute` int(11) unsigned NOT NULL DEFAULT '0',
--   `estimate_comment` varchar(255) NOT NULL DEFAULT '',
--   PRIMARY KEY (`project_stat_phase_id`),
--   CONSTRAINT `fk_project_stat_phase_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`),
--   CONSTRAINT `fk_project_stat_phase_phase` FOREIGN KEY (`stat_phase_id`) REFERENCES `stat_phase` (`stat_phase_id`),
--   CONSTRAINT `fk_project_stat_phase_bucket` FOREIGN KEY (`bucket_id`) REFERENCES `bucket` (`bucket_id`),
--   CONSTRAINT `fk_project_stat_phase_method` FOREIGN KEY (`method_id`) REFERENCES `method` (`method_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `phase_products_check`;
-- 
-- CREATE TABLE `phase_products_check` ( 
--   `phase_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `satisfied` tinyint(1) NOT NULL,
--   PRIMARY KEY (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- -- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `language`;
-- 
-- CREATE TABLE `language` (
--   `language_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `name` varchar(255) NOT NULL,
--   PRIMARY KEY (`language_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `design_method`;
-- 
-- CREATE TABLE `design_method` (
--   `design_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL, <---- the various design templates used
--   PRIMARY KEY (`language_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `stats_bucket`;
-- 
-- CREATE TABLE `project` (
--   `stats_bucket_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   PRIMARY KEY (`stats_bucket_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 

  -- process table
  -- `size_unit` varchar(255) NOT NULL, # LOC, word, etc.
  -- `size_k_unit` varchar(255) NOT NULL, # KLOC, per thousand words, etc.
  -- `entry_criteria` varchar(255) NOT NULL,
  -- `exit_criteria` varchar(255) NOT NULL,


-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `project`;
-- 
-- CREATE TABLE `project` (
--   `project_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `module_template_id` int(11) unsigned NOT NULL,
--   `project_bucket_id` int(11) unsigned,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   `current_phase_id` int(11) unsigned, # <--- indicates which forms are currently open
--   `current_subphase_id` int(11) unsigned, # <--- indicates which step of planning is taking place
--   `design_method_id` int(11) unsigned NOT NULL,
--   `size_estimation_method_id` int(11) unsigned NOT NULL,
--   `time_estimation_method_id` int(11) unsigned NOT NULL,
--   `design_method_id` int(11) unsigned NOT NULL,
--   `language_id` int(11) unsigned NOT NULL,
--   `mutli_day` tinyint(1) unsigned NOT NULL,
--   `planned_time` int(11) unsigned NOT NULL,
--   `actual_time` int(11) unsigned NOT NULL,
--   `planned_pct_reuse` int(11) unsigned NOT NULL,
--   `actual_pct_reuse` int(11) unsigned NOT NULL,
--   `planned_defect_count` int(11) unsigned NOT NULL,
--   `planned_appraisal_coq` int(11) unsigned NOT NULL,
--   `planned_failure_coq` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`project_id`),
--   CONSTRAINT `fk_project_process` FOREIGN KEY (`process_id`) REFERENCES `process` (`process_id`),
--   CONSTRAINT `fk_project_design` FOREIGN KEY (`design_method_id`) REFERENCES `language` (`language_id`),
--   CONSTRAINT `fk_project_language` FOREIGN KEY (`language_id`) REFERENCES `language` (`language_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `module_template`; # was project
-- 
-- CREATE TABLE `module_template` (
--   `module_template_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `process_id` int(11) unsigned,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   `process_id` int(11) unsigned,
--   `size_estimation_method_id` int(11) unsigned NOT NULL,
--   `time_estimation_method_id` int(11) unsigned NOT NULL,
--   `design_method_id` int(11) unsigned NOT NULL,
--   `language_id` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`project_id`),
--   CONSTRAINT `fk_project_process` FOREIGN KEY (`process_id`) REFERENCES `process` (`process_id`),
--   CONSTRAINT `fk_project_design` FOREIGN KEY (`design_method_id`) REFERENCES `language` (`language_id`),
--   CONSTRAINT `fk_project_language` FOREIGN KEY (`language_id`) REFERENCES `language` (`language_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `project_part`; # A project can have multiple language specific parts
-- 
-- CREATE TABLE `project_part` (
--   `project_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `module_template_id` int(11) unsigned NOT NULL,
--   `project_bucket_id` int(11) unsigned,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   `current_phase_id` int(11) unsigned, # <--- indicates which forms are currently open
--   `current_subphase_id` int(11) unsigned, # <--- indicates which step of planning is taking place
--   `design_method_id` int(11) unsigned NOT NULL,
--   `size_estimation_method_id` int(11) unsigned NOT NULL,
--   `time_estimation_method_id` int(11) unsigned NOT NULL,
--   `design_method_id` int(11) unsigned NOT NULL,
--   `language_id` int(11) unsigned NOT NULL,
--   `mutli_day` tinyint(1) unsigned NOT NULL,
--   `planned_time` int(11) unsigned NOT NULL,
--   `actual_time` int(11) unsigned NOT NULL,
--   `planned_pct_reuse` int(11) unsigned NOT NULL,
--   `actual_pct_reuse` int(11) unsigned NOT NULL,
--   `planned_defect_count` int(11) unsigned NOT NULL,
--   `planned_appraisal_coq` int(11) unsigned NOT NULL,
--   `planned_failure_coq` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`project_id`),
--   CONSTRAINT `fk_project_process` FOREIGN KEY (`process_id`) REFERENCES `process` (`process_id`),
--   CONSTRAINT `fk_project_design` FOREIGN KEY (`design_method_id`) REFERENCES `language` (`language_id`),
--   CONSTRAINT `fk_project_language` FOREIGN KEY (`language_id`) REFERENCES `language` (`language_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `project_cycle_plan`; # The planned values for recording are form a cycle.
-- 
-- CREATE TABLE `project_cycle` (
--   `project_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `module_template_id` int(11) unsigned NOT NULL,
--   `num` int(11) unsigned,
--   `planned_pct_reuse` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`project_id`),
--   CONSTRAINT `fk_project_process` FOREIGN KEY (`process_id`) REFERENCES `process` (`process_id`),
--   CONSTRAINT `fk_project_design` FOREIGN KEY (`design_method_id`) REFERENCES `language` (`language_id`),
--   CONSTRAINT `fk_project_language` FOREIGN KEY (`language_id`) REFERENCES `language` (`language_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `project_cycle_actual`; # The actual values for recording are form a cycle.
-- 
-- CREATE TABLE `project_cycle` (
--   `project_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `module_template_id` int(11) unsigned NOT NULL,
--   `num` int(11) unsigned,
--   `actual_pct_reuse` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`project_id`),
--   CONSTRAINT `fk_project_process` FOREIGN KEY (`process_id`) REFERENCES `process` (`process_id`),
--   CONSTRAINT `fk_project_design` FOREIGN KEY (`design_method_id`) REFERENCES `language` (`language_id`),
--   CONSTRAINT `fk_project_language` FOREIGN KEY (`language_id`) REFERENCES `language` (`language_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `estimate_loc`;
-- 
-- CREATE TABLE `estimate_loc` (
--   `estimate_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `base` int(11) unsigned NOT NULL,
--   `new` int(11) unsigned NOT NULL,
--   `changed` int(11) unsigned NOT NULL,
--   `added` int(11) unsigned NOT NULL,
--   `modified` int(11) unsigned NOT NULL,
--   `deleted` int(11) unsigned NOT NULL,
--   `reused` int(11) unsigned NOT NULL,
--   `object_loc` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`estimate_id`),
--   UNIQUE KEY (`project_id`),
--   CONSTRAINT `fk_estimate_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`),
--   CONSTRAINT `fk_estimate_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `probe_object_type`;
-- 
-- CREATE TABLE `probe_object_type` (
--   `estimate_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `base_loc` int(11) unsigned NOT NULL,
--   `deleted_loc` int(11) unsigned NOT NULL,
--   `modified_loc` int(11) unsigned NOT NULL,
--   `b0_size` int(11) unsigned NOT NULL,
--   `b1_size` int(11) unsigned NOT NULL,
--   `b0_time` int(11) unsigned NOT NULL,
--   `b1_time` int(11) unsigned NOT NULL,
--   `new_loc` int(11) unsigned NOT NULL,
--   `new_reuse_loc` int(11) unsigned NOT NULL,
--   `estimated_time_min` int(11) unsigned NOT NULL,
--   `upper_prediction_interval` int(11) unsigned NOT NULL,
--   `lower_prediction_interval` int(11) unsigned NOT NULL,
--   `prediction_interval_percent` int(11) unsigned NOT NULL,
--   `acual_base_loc` int(11) unsigned NOT NULL,
--   `acual_deleted_loc` int(11) unsigned NOT NULL,
--   `acual_modified_loc` int(11) unsigned NOT NULL,
--   `loc_upper_prediction_interval_70` int(11) unsigned NOT NULL,
--   `loc_lower_prediction_interval_70` int(11) unsigned NOT NULL,
--   `time_upper_prediction_interval_70` int(11) unsigned NOT NULL,
--   `time_lower_prediction_interval_70` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`estimate_id`),
--   UNIQUE KEY (`project_id`, `phase_id`),
--   CONSTRAINT `fk_estimate_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`),
--   CONSTRAINT `fk_estimate_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `estimate_probe`;
-- 
-- CREATE TABLE `estimate_probe` (
--   `estimate_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `base_loc` int(11) unsigned NOT NULL,
--   `deleted_loc` int(11) unsigned NOT NULL,
--   `modified_loc` int(11) unsigned NOT NULL,
--   `b0_size` int(11) unsigned NOT NULL,
--   `b1_size` int(11) unsigned NOT NULL,
--   `b0_time` int(11) unsigned NOT NULL,
--   `b1_time` int(11) unsigned NOT NULL,
--   `new_loc` int(11) unsigned NOT NULL,
--   `new_reuse_loc` int(11) unsigned NOT NULL,
--   `estimated_time_min` int(11) unsigned NOT NULL,
--   `upper_prediction_interval` int(11) unsigned NOT NULL,
--   `lower_prediction_interval` int(11) unsigned NOT NULL,
--   `prediction_interval_percent` int(11) unsigned NOT NULL,
--   `acual_base_loc` int(11) unsigned NOT NULL,
--   `acual_deleted_loc` int(11) unsigned NOT NULL,
--   `acual_modified_loc` int(11) unsigned NOT NULL,
--   `loc_upper_prediction_interval_70` int(11) unsigned NOT NULL,
--   `loc_lower_prediction_interval_70` int(11) unsigned NOT NULL,
--   `time_upper_prediction_interval_70` int(11) unsigned NOT NULL,
--   `time_lower_prediction_interval_70` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`estimate_id`),
--   UNIQUE KEY (`project_id`, `phase_id`),
--   CONSTRAINT `fk_estimate_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`),
--   CONSTRAINT `fk_estimate_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `probe_type`;
-- 
-- CREATE TABLE `probe_type` (
--   `probe_type_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `number` int(11) unsigned NOT NULL,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   PRIMARY KEY (`defect_type_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `probe_object_size`;
-- 
-- CREATE TABLE `probe_type` (
--   `probe_object_size` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `number` int(11) unsigned NOT NULL,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   PRIMARY KEY (`defect_type_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `estimate_probe_add_loc`;
-- 
-- CREATE TABLE `estimate_probe_add_loc` (
--   `estimate_probe_add_loc_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `name` varchar(255) NOT NULL,
--   `type` int(11) unsigned NOT NULL,
--   `relative_size` int(11) unsigned NOT NULL,
--   `loc` int(11) unsigned NOT NULL,
--   `acual_loc` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`estimate_id`),
--   UNIQUE KEY (`project_id`, `phase_id`),
--   CONSTRAINT `fk_estimate_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`),
--   CONSTRAINT `fk_estimate_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `estimate_probe_object_loc`;
-- 
-- CREATE TABLE `estimate_probe_object_loc` (
--   `estimate_probe_add_loc_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `name` varchar(255) NOT NULL,
--   `type` int(11) unsigned NOT NULL,
--   `relative_size` int(11) unsigned NOT NULL,
--   `loc_per_method` int(11) unsigned NOT NULL,
--   `acual_loc` int(11) unsigned NOT NULL,
--   `for_reuse` tinyint(1) unsigned NOT NULL,
--   PRIMARY KEY (`estimate_id`),
--   UNIQUE KEY (`project_id`, `phase_id`),
--   CONSTRAINT `fk_estimate_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`),
--   CONSTRAINT `fk_estimate_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `estimate_probe_object_reused`;
-- 
-- CREATE TABLE `estimate_probe` (
--   `estimate_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `name` varchar(255) NOT NULL,
--   `loc` int(11) unsigned NOT NULL,
--   `acual_loc` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`estimate_id`),
--   UNIQUE KEY (`project_id`, `phase_id`),
--   CONSTRAINT `fk_estimate_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`),
--   CONSTRAINT `fk_estimate_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `actual_loc`;
-- 
-- CREATE TABLE `actual_loc` (
--   `estimate_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `cycle` int(11) unsigned NOT NULL,
--   `base` int(11) unsigned NOT NULL,
--   `new` int(11) unsigned NOT NULL,
--   `changed` int(11) unsigned NOT NULL,
--   `added` int(11) unsigned NOT NULL,
--   `modified` int(11) unsigned NOT NULL,
--   `deleted` int(11) unsigned NOT NULL,
--   `reused` int(11) unsigned NOT NULL,
--   `object_loc` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`estimate_id`),
--   UNIQUE KEY (`project_id`),
--   CONSTRAINT `fk_estimate_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`),
--   CONSTRAINT `fk_estimate_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `schedule`;
-- 
-- CREATE TABLE `schedule` (
--   `schedule_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `day_or_week` tinyint(1) unsigned NOT NULL,
--   PRIMARY KEY (`defect_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `schedule_week`;
-- 
-- CREATE TABLE `schedule_week` (
--   `schedule_week_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `schedule_id` int(11) unsigned NOT NULL,
--   `project_id` int(11) unsigned NOT NULL,
--   `num` int(11) unsigned NOT NULL,
--   `date_monday` datetime NOT NULL,
--   PRIMARY KEY (`defect_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `task`;
-- 
-- CREATE TABLE `task` (
--   `task_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `phase_id` int(11) unsigned NOT NULL,
--   `schedule_week_id` int(11) unsigned NOT NULL,
--   `num` int(11) unsigned NOT NULL,
--   `name` varchar(255) NOT NULL,
--   `planned_hours` int(11) unsigned NOT NULL,
--   `pct_complete` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`defect_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `time_log`;
-- 
-- CREATE TABLE `time_log` (
--   `time_log_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `project_id` int(11) unsigned NOT NULL,
--   `phase_id` int(11) unsigned NOT NULL,
--   `cycle` int(11) unsigned NOT NULL,
--   `task_id` int(11) unsigned NOT NULL,
--   `start_time` DATETIME NOT NULL,
--   `stop_time` DATETIME NOT NULL, 
--   `interruption_minutes` int(11) unsigned NOT NULL,
--   `comments` varchar(255) NOT NULL,
--   PRIMARY KEY (`time_log_id`),
--   CONSTRAINT `fk_time_log_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`project_id`)
--   CONSTRAINT `fk_time_log_phase` FOREIGN KEY (`phase_id`) REFERENCES `phase` (`phase_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `defect_type`;
-- 
-- CREATE TABLE `defect_type` (
--   `defect_type_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `number` int(11) unsigned NOT NULL,
--   `name` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   PRIMARY KEY (`defect_type_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `defect`;
-- 
-- CREATE TABLE `defect` (
--   `defect_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `found_time` DATETIME NOT NULL,
--   `inject_project_id` int(11) unsigned NOT NULL,
--   `cycle` int(11) unsigned NOT NULL,
--   `inject_phase_id` int(11) unsigned NOT NULL,
--   `cycle` int(11) unsigned NOT NULL,
--   `remove_project_id` int(11) unsigned NOT NULL,
--   `remove_phase_id` int(11) unsigned NOT NULL,
--   `fix_minutes` int(11) unsigned NOT NULL,
--   `source_defect_id` int(11) unsigned NOT NULL,
--   `description` varchar(255) NOT NULL,
--   PRIMARY KEY (`defect_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `issue`;
-- 
-- CREATE TABLE `issue` (
--   `issue_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `found_time` DATETIME NOT NULL,
--   `inject_project_id` int(11) unsigned NOT NULL,
--   `cycle` int(11) unsigned NOT NULL,
--   `inject_phase_id` int(11) unsigned NOT NULL,
--   `description` varchar(255) NOT NULL,
--   `resolution_time` DATETIME NOT NULL,
--   `resolution` varchar(255) NOT NULL,
--   PRIMARY KEY (`defect_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `pip`;
-- 
-- CREATE TABLE `pip` (
--   `pip_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `found_time` DATETIME NOT NULL,
--   `project_id` int(11) unsigned NOT NULL,
--   `process_id` int(11) unsigned NOT NULL,
--   `phase_id` int(11) unsigned NOT NULL,
--   `subphase_id` int(11) unsigned NOT NULL,
--   `problem` varchar(255) NOT NULL,
--   `proposal` varchar(255) NOT NULL,
--   `resolved_time` DATETIME NOT NULL,
--   `resolved_in_process_id` int(11) unsigned NOT NULL,
--   PRIMARY KEY (`defect_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `test_case`;
-- 
-- CREATE TABLE `test_case` (
--   `test_case_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `found_time` DATETIME NOT NULL,
--   `project_id` int(11) unsigned NOT NULL,
--   `objective` varchar(255) NOT NULL,
--   `description` varchar(255) NOT NULL,
--   `conditions` varchar(255) NOT NULL,
--   `expected` varchar(255) NOT NULL,
--   PRIMARY KEY (`defect_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- 
-- # ------------------------------------------------------------
-- 
-- DROP TABLE IF EXISTS `test_case_result`;
-- 
-- CREATE TABLE `test_case_result` (
--   `test_case_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
--   `run_time` DATETIME NOT NULL,
--   `project_id` int(11) unsigned NOT NULL,
--   `actual` varchar(255) NOT NULL,
--   PRIMARY KEY (`defect_id`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
