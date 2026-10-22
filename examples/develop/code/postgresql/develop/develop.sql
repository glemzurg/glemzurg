-- Development process and project tracking.
-- One row is one object from the development requirements model.
-- Owned children cascade on delete. A reference to a definition restricts delete.

--------------------------------------------------------------

CREATE TYPE process_state AS ENUM ('editable', 'locked');
COMMENT ON TYPE process_state IS 'Whether a process can still be changed.';

CREATE TYPE stage_type AS ENUM ('once', 'repeating');
COMMENT ON TYPE stage_type IS 'Whether a process stage runs once or repeats.';

CREATE TYPE estimate_method AS ENUM ('guess', 'computed');
COMMENT ON TYPE estimate_method IS 'Which method specialization an estimate uses.';

--------------------------------------------------------------

CREATE TABLE family (
  family_key text NOT NULL,
  name text NOT NULL,
  description text DEFAULT NULL,
  PRIMARY KEY (family_key)
);

COMMENT ON TABLE family IS 'A shared group of processes, and the projects that use those processes.';
COMMENT ON COLUMN family.family_key IS 'The internal ID.';
COMMENT ON COLUMN family.name IS 'The name of this family.';
COMMENT ON COLUMN family.description IS 'A description of this family of processes.';

--------------------------------------------------------------

CREATE TABLE phase (
  phase_key text NOT NULL,
  family_key text NOT NULL,
  num integer NOT NULL,
  name text NOT NULL,
  description text DEFAULT NULL,
  PRIMARY KEY (phase_key),
  UNIQUE (family_key, phase_key),
  UNIQUE (family_key, num),
  UNIQUE (family_key, name),
  CONSTRAINT fk_phase_family FOREIGN KEY (family_key) REFERENCES family (family_key) ON DELETE CASCADE,
  CONSTRAINT chk_phase_num CHECK (num >= 1)
);

COMMENT ON TABLE phase IS 'Fundamental phase skeleton shared by every process in a family.';
COMMENT ON COLUMN phase.phase_key IS 'The internal ID.';
COMMENT ON COLUMN phase.family_key IS 'The family this phase belongs to.';
COMMENT ON COLUMN phase.num IS 'Order of this phase within the family. The same order is used across that family''s processes.';
COMMENT ON COLUMN phase.name IS 'Name of this phase. Unique among phases of the same family.';
COMMENT ON COLUMN phase.description IS 'A high level description of this phase.';

--------------------------------------------------------------

CREATE TABLE defect_type (
  defect_type_key text NOT NULL,
  family_key text NOT NULL,
  parent_defect_type_key text NOT NULL,
  num integer NOT NULL,
  name text NOT NULL,
  description text DEFAULT NULL,
  PRIMARY KEY (defect_type_key),
  UNIQUE (family_key, num),
  UNIQUE (family_key, name),
  CONSTRAINT fk_defect_type_family FOREIGN KEY (family_key) REFERENCES family (family_key) ON DELETE CASCADE,
  CONSTRAINT fk_defect_type_parent FOREIGN KEY (parent_defect_type_key) REFERENCES defect_type (defect_type_key) ON DELETE RESTRICT,
  CONSTRAINT chk_defect_type_num CHECK (num >= 10)
);

COMMENT ON TABLE defect_type IS 'A type of defect classified within a process family.';
COMMENT ON COLUMN defect_type.defect_type_key IS 'The internal ID.';
COMMENT ON COLUMN defect_type.family_key IS 'The family this defect type belongs to.';
COMMENT ON COLUMN defect_type.parent_defect_type_key IS 'The defect type this one is a subtype of. Every defect type has exactly one.';
COMMENT ON COLUMN defect_type.num IS 'The unique value for this defect type. Every tenth value is a base type, and the numbers up to the next base are subtypes.';
COMMENT ON COLUMN defect_type.name IS 'A simple name for this type of defect. Unique among defect types of the same family.';
COMMENT ON COLUMN defect_type.description IS 'What kind of defects are this type.';

--------------------------------------------------------------

CREATE TABLE size_count_type (
  size_count_type_key text NOT NULL,
  family_key text NOT NULL,
  name text NOT NULL,
  description text DEFAULT NULL,
  PRIMARY KEY (size_count_type_key),
  UNIQUE (family_key, name),
  CONSTRAINT fk_size_count_type_family FOREIGN KEY (family_key) REFERENCES family (family_key) ON DELETE CASCADE
);

COMMENT ON TABLE size_count_type IS 'A named kind of count for size.';
COMMENT ON COLUMN size_count_type.size_count_type_key IS 'The internal ID.';
COMMENT ON COLUMN size_count_type.family_key IS 'The family this size count belongs to.';
COMMENT ON COLUMN size_count_type.name IS 'The name of this size count. Unique among size counts of the same family.';
COMMENT ON COLUMN size_count_type.description IS 'What this size count measures.';

--------------------------------------------------------------

CREATE TABLE size_category (
  size_category_key text NOT NULL,
  size_count_type_key text NOT NULL,
  num integer NOT NULL,
  name text NOT NULL,
  description text DEFAULT NULL,
  PRIMARY KEY (size_category_key),
  UNIQUE (size_count_type_key, num),
  UNIQUE (size_count_type_key, name),
  CONSTRAINT fk_size_category_count FOREIGN KEY (size_count_type_key) REFERENCES size_count_type (size_count_type_key) ON DELETE CASCADE,
  CONSTRAINT chk_size_category_num CHECK (num >= 0)
);

COMMENT ON TABLE size_category IS 'An ordered category of a size count type.';
COMMENT ON COLUMN size_category.size_category_key IS 'The internal ID.';
COMMENT ON COLUMN size_category.size_count_type_key IS 'The size count this category belongs to.';
COMMENT ON COLUMN size_category.num IS 'Order of this category within the size count type.';
COMMENT ON COLUMN size_category.name IS 'Name of this category. Unique among categories of the same size count type.';
COMMENT ON COLUMN size_category.description IS 'What this category covers.';

--------------------------------------------------------------

CREATE TABLE process (
  process_key text NOT NULL,
  family_key text NOT NULL,
  ancestor_process_key text DEFAULT NULL,
  name text NOT NULL,
  version integer NOT NULL,
  version_minor integer NOT NULL DEFAULT 0,
  purpose text NOT NULL,
  entry_criteria text NOT NULL,
  exit_criteria text NOT NULL,
  state process_state NOT NULL DEFAULT 'editable',
  PRIMARY KEY (process_key),
  UNIQUE (family_key, process_key),
  UNIQUE (family_key, name, version, version_minor),
  CONSTRAINT fk_process_family FOREIGN KEY (family_key) REFERENCES family (family_key) ON DELETE CASCADE,
  CONSTRAINT fk_process_ancestor FOREIGN KEY (ancestor_process_key) REFERENCES process (process_key) ON DELETE RESTRICT,
  CONSTRAINT chk_process_version CHECK (version >= 0),
  CONSTRAINT chk_process_version_minor CHECK (version_minor >= 0)
);

COMMENT ON TABLE process IS 'A versioned process to follow, owned by a family.';
COMMENT ON COLUMN process.process_key IS 'The internal ID.';
COMMENT ON COLUMN process.family_key IS 'The family this process belongs to.';
COMMENT ON COLUMN process.ancestor_process_key IS 'Earlier process this version replaces, when one exists.';
COMMENT ON COLUMN process.name IS 'Name of this process. Unique together with version and version minor within the family.';
COMMENT ON COLUMN process.version IS 'Major version of this process.';
COMMENT ON COLUMN process.version_minor IS 'Minor version of this process.';
COMMENT ON COLUMN process.purpose IS 'Why this process is followed.';
COMMENT ON COLUMN process.entry_criteria IS 'What must be true before this process starts.';
COMMENT ON COLUMN process.exit_criteria IS 'What must be true before this process is done.';
COMMENT ON COLUMN process.state IS 'Editable after creation. Lock moves the process to locked, after which it is no longer changed.';

--------------------------------------------------------------

CREATE TABLE stage (
  stage_key text NOT NULL,
  process_key text NOT NULL,
  family_key text NOT NULL,
  phase_key text NOT NULL,
  num integer NOT NULL,
  name text NOT NULL,
  task_summary text NOT NULL,
  purpose text NOT NULL,
  entry_criteria text NOT NULL,
  exit_criteria text NOT NULL,
  type stage_type NOT NULL,
  PRIMARY KEY (stage_key),
  UNIQUE (process_key, num),
  UNIQUE (process_key, name),
  CONSTRAINT fk_stage_process FOREIGN KEY (family_key, process_key) REFERENCES process (family_key, process_key) ON DELETE CASCADE,
  CONSTRAINT fk_stage_phase FOREIGN KEY (family_key, phase_key) REFERENCES phase (family_key, phase_key) ON DELETE RESTRICT,
  CONSTRAINT chk_stage_num CHECK (num >= 0)
);

COMMENT ON TABLE stage IS 'A set of scripts within a process. A stage runs once or is repeating.';
COMMENT ON COLUMN stage.stage_key IS 'The internal ID.';
COMMENT ON COLUMN stage.process_key IS 'The process this stage belongs to.';
COMMENT ON COLUMN stage.family_key IS 'The family of the owning process. Stored so the stage and its family phase are the same family.';
COMMENT ON COLUMN stage.phase_key IS 'Family phase this stage is performed in.';
COMMENT ON COLUMN stage.num IS 'Order of this stage within the process.';
COMMENT ON COLUMN stage.name IS 'Name of this stage. Unique among stages of the same process.';
COMMENT ON COLUMN stage.task_summary IS 'Summary of the work in this stage.';
COMMENT ON COLUMN stage.purpose IS 'Why this stage is performed.';
COMMENT ON COLUMN stage.entry_criteria IS 'What must be true before this stage starts.';
COMMENT ON COLUMN stage.exit_criteria IS 'What must be true before this stage is done.';
COMMENT ON COLUMN stage.type IS 'Whether this stage runs once or is repeating.';

--------------------------------------------------------------

CREATE TABLE script (
  script_key text NOT NULL,
  stage_key text NOT NULL,
  num integer NOT NULL,
  name text NOT NULL,
  task_summary text NOT NULL,
  purpose text NOT NULL,
  entry_criteria text NOT NULL,
  exit_criteria text NOT NULL,
  PRIMARY KEY (script_key),
  UNIQUE (stage_key, num),
  UNIQUE (stage_key, name),
  CONSTRAINT fk_script_stage FOREIGN KEY (stage_key) REFERENCES stage (stage_key) ON DELETE CASCADE,
  CONSTRAINT chk_script_num CHECK (num >= 0)
);

COMMENT ON TABLE script IS 'A step-by-step process script owned by a stage.';
COMMENT ON COLUMN script.script_key IS 'The internal ID.';
COMMENT ON COLUMN script.stage_key IS 'The stage this script belongs to.';
COMMENT ON COLUMN script.num IS 'Order of this script within the stage.';
COMMENT ON COLUMN script.name IS 'Name of this script. Unique among scripts of the same stage.';
COMMENT ON COLUMN script.task_summary IS 'Summary of the work in this script.';
COMMENT ON COLUMN script.purpose IS 'Why this script is followed.';
COMMENT ON COLUMN script.entry_criteria IS 'What must be true before this script starts.';
COMMENT ON COLUMN script.exit_criteria IS 'What must be true before this script is done.';

--------------------------------------------------------------

CREATE TABLE step (
  step_key text NOT NULL,
  script_key text NOT NULL,
  num integer NOT NULL,
  name text NOT NULL,
  PRIMARY KEY (step_key),
  UNIQUE (script_key, num),
  UNIQUE (script_key, name),
  CONSTRAINT fk_step_script FOREIGN KEY (script_key) REFERENCES script (script_key) ON DELETE CASCADE,
  CONSTRAINT chk_step_num CHECK (num >= 0)
);

COMMENT ON TABLE step IS 'A step of a process script.';
COMMENT ON COLUMN step.step_key IS 'The internal ID.';
COMMENT ON COLUMN step.script_key IS 'The script this step belongs to.';
COMMENT ON COLUMN step.num IS 'Order of this step within the script.';
COMMENT ON COLUMN step.name IS 'Name of this step. Unique among steps of the same script.';

--------------------------------------------------------------

CREATE TABLE step_task (
  step_task_key text NOT NULL,
  step_key text NOT NULL,
  num integer NOT NULL,
  name text NOT NULL,
  description text NOT NULL,
  PRIMARY KEY (step_task_key),
  UNIQUE (step_key, num),
  CONSTRAINT fk_step_task_step FOREIGN KEY (step_key) REFERENCES step (step_key) ON DELETE CASCADE,
  CONSTRAINT chk_step_task_num CHECK (num >= 0)
);

COMMENT ON TABLE step_task IS 'Work defined on a process step and shared by every project that follows the process.';
COMMENT ON COLUMN step_task.step_task_key IS 'The internal ID.';
COMMENT ON COLUMN step_task.step_key IS 'The step this task belongs to. A step has one or more step tasks.';
COMMENT ON COLUMN step_task.num IS 'Order of this step task within the step.';
COMMENT ON COLUMN step_task.name IS 'Name of this step task.';
COMMENT ON COLUMN step_task.description IS 'Work performed for this step task.';

--------------------------------------------------------------

CREATE TABLE project (
  project_key text NOT NULL,
  process_key text NOT NULL,
  name text NOT NULL,
  description text DEFAULT NULL,
  created_time timestamptz NOT NULL,
  PRIMARY KEY (project_key),
  CONSTRAINT fk_project_process FOREIGN KEY (process_key) REFERENCES process (process_key) ON DELETE RESTRICT
);

COMMENT ON TABLE project IS 'Work that follows one process.';
COMMENT ON COLUMN project.project_key IS 'The internal ID.';
COMMENT ON COLUMN project.process_key IS 'The process this project follows.';
COMMENT ON COLUMN project.name IS 'The name of this project.';
COMMENT ON COLUMN project.description IS 'A description of this project.';
COMMENT ON COLUMN project.created_time IS 'When this project was created.';

--------------------------------------------------------------

CREATE TABLE cycle (
  cycle_key text NOT NULL,
  project_key text NOT NULL,
  stage_key text NOT NULL,
  num integer NOT NULL,
  PRIMARY KEY (cycle_key),
  UNIQUE (project_key, num),
  CONSTRAINT fk_cycle_project FOREIGN KEY (project_key) REFERENCES project (project_key) ON DELETE CASCADE,
  CONSTRAINT fk_cycle_stage FOREIGN KEY (stage_key) REFERENCES stage (stage_key) ON DELETE RESTRICT,
  CONSTRAINT chk_cycle_num CHECK (num >= 0)
);

COMMENT ON TABLE cycle IS 'A run of one process stage within a project.';
COMMENT ON COLUMN cycle.cycle_key IS 'The internal ID.';
COMMENT ON COLUMN cycle.project_key IS 'The project this cycle belongs to.';
COMMENT ON COLUMN cycle.stage_key IS 'The process stage this cycle is an instance of.';
COMMENT ON COLUMN cycle.num IS 'Order of this cycle within the project.';

--------------------------------------------------------------

CREATE TABLE step_instance (
  step_instance_key text NOT NULL,
  cycle_key text NOT NULL,
  step_key text NOT NULL,
  num integer NOT NULL,
  PRIMARY KEY (step_instance_key),
  UNIQUE (cycle_key, num),
  CONSTRAINT fk_step_instance_cycle FOREIGN KEY (cycle_key) REFERENCES cycle (cycle_key) ON DELETE CASCADE,
  CONSTRAINT fk_step_instance_step FOREIGN KEY (step_key) REFERENCES step (step_key) ON DELETE RESTRICT,
  CONSTRAINT chk_step_instance_num CHECK (num >= 0)
);

COMMENT ON TABLE step_instance IS 'Work in a project cycle. An instance of a process step.';
COMMENT ON COLUMN step_instance.step_instance_key IS 'The internal ID.';
COMMENT ON COLUMN step_instance.cycle_key IS 'The cycle this step instance belongs to.';
COMMENT ON COLUMN step_instance.step_key IS 'The process step this step instance is an instance of.';
COMMENT ON COLUMN step_instance.num IS 'Order of this step instance within the cycle.';

--------------------------------------------------------------

CREATE TABLE task (
  task_key text NOT NULL,
  step_instance_key text NOT NULL,
  step_task_key text NOT NULL,
  planned_hours integer NOT NULL,
  pct_complete integer NOT NULL,
  PRIMARY KEY (task_key),
  CONSTRAINT fk_task_step_instance FOREIGN KEY (step_instance_key) REFERENCES step_instance (step_instance_key) ON DELETE CASCADE,
  CONSTRAINT fk_task_step_task FOREIGN KEY (step_task_key) REFERENCES step_task (step_task_key) ON DELETE RESTRICT,
  CONSTRAINT chk_task_planned_hours CHECK (planned_hours >= 0),
  CONSTRAINT chk_task_pct_complete CHECK (pct_complete >= 0 AND pct_complete <= 100)
);

COMMENT ON TABLE task IS 'A project task. An instance of a process step task.';
COMMENT ON COLUMN task.task_key IS 'The internal ID.';
COMMENT ON COLUMN task.step_instance_key IS 'The step instance this task belongs to.';
COMMENT ON COLUMN task.step_task_key IS 'The process step task this task is an instance of.';
COMMENT ON COLUMN task.planned_hours IS 'Planned hours for this task, in whole hours.';
COMMENT ON COLUMN task.pct_complete IS 'Percent complete, from 0 to 100.';

--------------------------------------------------------------

CREATE TABLE estimate (
  estimate_key text NOT NULL,
  size_category_key text NOT NULL,
  task_key text DEFAULT NULL,
  method estimate_method NOT NULL,
  mean numeric NOT NULL,
  variance numeric NOT NULL,
  low numeric NOT NULL,
  high numeric NOT NULL,
  prediction_interval numeric NOT NULL,
  comment text NOT NULL,
  estimation_time timestamptz NOT NULL,
  PRIMARY KEY (estimate_key),
  UNIQUE (task_key),
  CONSTRAINT fk_estimate_size_category FOREIGN KEY (size_category_key) REFERENCES size_category (size_category_key) ON DELETE RESTRICT,
  CONSTRAINT fk_estimate_task FOREIGN KEY (task_key) REFERENCES task (task_key) ON DELETE CASCADE,
  CONSTRAINT chk_estimate_variance CHECK (variance >= 0),
  CONSTRAINT chk_estimate_prediction_interval CHECK (prediction_interval >= 0 AND prediction_interval <= 1)
);

COMMENT ON TABLE estimate IS 'An estimate of time, organized and later closed by the size category it uses.';
COMMENT ON COLUMN estimate.estimate_key IS 'The internal ID.';
COMMENT ON COLUMN estimate.size_category_key IS 'Size category this estimate is categorized by.';
COMMENT ON COLUMN estimate.task_key IS 'Task this size estimate belongs to, when one is set. At most one estimate per task.';
COMMENT ON COLUMN estimate.method IS 'Which specialization holds the rest of this estimate. Exactly one of guess or computed.';
COMMENT ON COLUMN estimate.mean IS 'The middle of the estimate bell curve.';
COMMENT ON COLUMN estimate.variance IS 'The shape of the estimate bell curve.';
COMMENT ON COLUMN estimate.low IS 'Lowest estimate based on the prediction interval.';
COMMENT ON COLUMN estimate.high IS 'Highest estimate based on the prediction interval.';
COMMENT ON COLUMN estimate.prediction_interval IS 'High and low range of interest, from 0 to 1.';
COMMENT ON COLUMN estimate.comment IS 'Note about how the estimate was made.';
COMMENT ON COLUMN estimate.estimation_time IS 'When the estimate was made.';

--------------------------------------------------------------

CREATE TABLE guess (
  estimate_key text NOT NULL,
  guess_lowest_80pred integer NOT NULL,
  guess_highest_80pred integer NOT NULL,
  PRIMARY KEY (estimate_key),
  CONSTRAINT fk_guess_estimate FOREIGN KEY (estimate_key) REFERENCES estimate (estimate_key) ON DELETE CASCADE,
  CONSTRAINT chk_guess_lowest CHECK (guess_lowest_80pred >= 0),
  CONSTRAINT chk_guess_highest CHECK (guess_highest_80pred >= 0)
);

COMMENT ON TABLE guess IS 'An estimate made as a direct guess.';
COMMENT ON COLUMN guess.estimate_key IS 'The estimate this guess specializes. The estimate method is guess.';
COMMENT ON COLUMN guess.guess_lowest_80pred IS 'Lowest expected value at an 80 percent prediction interval.';
COMMENT ON COLUMN guess.guess_highest_80pred IS 'Highest expected value at an 80 percent prediction interval.';

--------------------------------------------------------------

CREATE TABLE computed (
  estimate_key text NOT NULL,
  count integer NOT NULL,
  mean numeric NOT NULL,
  variance numeric NOT NULL,
  PRIMARY KEY (estimate_key),
  CONSTRAINT fk_computed_estimate FOREIGN KEY (estimate_key) REFERENCES estimate (estimate_key) ON DELETE CASCADE,
  CONSTRAINT chk_computed_count CHECK (count >= 0),
  CONSTRAINT chk_computed_variance CHECK (variance >= 0)
);

COMMENT ON TABLE computed IS 'An estimate computed from historic data.';
COMMENT ON COLUMN computed.estimate_key IS 'The estimate this computation specializes. The estimate method is computed.';
COMMENT ON COLUMN computed.count IS 'Number of values used in this computation.';
COMMENT ON COLUMN computed.mean IS 'Middle of the computed bell curve.';
COMMENT ON COLUMN computed.variance IS 'Shape of the computed bell curve.';

--------------------------------------------------------------

CREATE TABLE actual (
  actual_key text NOT NULL,
  estimate_key text NOT NULL,
  value integer NOT NULL,
  completion_time timestamptz NOT NULL,
  PRIMARY KEY (actual_key),
  UNIQUE (estimate_key),
  CONSTRAINT fk_actual_estimate FOREIGN KEY (estimate_key) REFERENCES estimate (estimate_key) ON DELETE CASCADE,
  CONSTRAINT chk_actual_value CHECK (value >= 0)
);

COMMENT ON TABLE actual IS 'The recorded result of an estimate. Added when the work is done, which completes the estimate.';
COMMENT ON COLUMN actual.actual_key IS 'The internal ID.';
COMMENT ON COLUMN actual.estimate_key IS 'The estimate this result completes. An estimate has at most one actual.';
COMMENT ON COLUMN actual.value IS 'Value recorded when the estimated work is complete, in whole units.';
COMMENT ON COLUMN actual.completion_time IS 'When the work was completed.';

--------------------------------------------------------------

CREATE TABLE defect (
  defect_key text NOT NULL,
  defect_type_key text NOT NULL,
  injected_task_key text NOT NULL,
  removed_step_instance_key text NOT NULL,
  source_defect_key text DEFAULT NULL,
  found_time timestamptz NOT NULL,
  cycle integer NOT NULL,
  fix_minutes integer NOT NULL,
  description text DEFAULT NULL,
  prevention text NOT NULL,
  test_defect boolean NOT NULL,
  PRIMARY KEY (defect_key),
  CONSTRAINT fk_defect_type FOREIGN KEY (defect_type_key) REFERENCES defect_type (defect_type_key) ON DELETE RESTRICT,
  CONSTRAINT fk_defect_injected_task FOREIGN KEY (injected_task_key) REFERENCES task (task_key) ON DELETE RESTRICT,
  CONSTRAINT fk_defect_removed_step FOREIGN KEY (removed_step_instance_key) REFERENCES step_instance (step_instance_key) ON DELETE RESTRICT,
  CONSTRAINT fk_defect_source FOREIGN KEY (source_defect_key) REFERENCES defect (defect_key) ON DELETE RESTRICT,
  CONSTRAINT chk_defect_cycle CHECK (cycle >= 0),
  CONSTRAINT chk_defect_fix_minutes CHECK (fix_minutes >= 0)
);

COMMENT ON TABLE defect IS 'A defect injected in a task and removed in a step instance.';
COMMENT ON COLUMN defect.defect_key IS 'The internal ID.';
COMMENT ON COLUMN defect.defect_type_key IS 'The category of defect this is.';
COMMENT ON COLUMN defect.injected_task_key IS 'Task where this defect was injected.';
COMMENT ON COLUMN defect.removed_step_instance_key IS 'Step instance where this defect was removed.';
COMMENT ON COLUMN defect.source_defect_key IS 'Defect this one was cloned from, when one exists.';
COMMENT ON COLUMN defect.found_time IS 'When this defect was found.';
COMMENT ON COLUMN defect.cycle IS 'Cycle number recorded for this defect.';
COMMENT ON COLUMN defect.fix_minutes IS 'Minutes spent fixing this defect.';
COMMENT ON COLUMN defect.description IS 'What the defect was.';
COMMENT ON COLUMN defect.prevention IS 'A prescriptive measure to catch this defect earlier or prevent it.';
COMMENT ON COLUMN defect.test_defect IS 'A defect in a test itself.';

--------------------------------------------------------------

CREATE TABLE time_log (
  time_log_key text NOT NULL,
  task_key text DEFAULT NULL,
  defect_key text DEFAULT NULL,
  cycle integer NOT NULL,
  start_time timestamptz NOT NULL,
  stop_time timestamptz NOT NULL,
  interruption_minutes integer NOT NULL,
  comments text NOT NULL,
  PRIMARY KEY (time_log_key),
  CONSTRAINT fk_time_log_task FOREIGN KEY (task_key) REFERENCES task (task_key) ON DELETE CASCADE,
  CONSTRAINT fk_time_log_defect FOREIGN KEY (defect_key) REFERENCES defect (defect_key) ON DELETE CASCADE,
  CONSTRAINT chk_time_log_one_parent CHECK ((task_key IS NOT NULL) <> (defect_key IS NOT NULL)),
  CONSTRAINT chk_time_log_cycle CHECK (cycle >= 0),
  CONSTRAINT chk_time_log_interruption CHECK (interruption_minutes >= 0),
  CONSTRAINT chk_time_log_interval CHECK (stop_time >= start_time)
);

COMMENT ON TABLE time_log IS 'A recorded interval of work on one task or one defect.';
COMMENT ON COLUMN time_log.time_log_key IS 'The internal ID.';
COMMENT ON COLUMN time_log.task_key IS 'Task this time was recorded against. Set when the log is not for a defect.';
COMMENT ON COLUMN time_log.defect_key IS 'Defect this time was recorded against. Set when the log is not for a task.';
COMMENT ON COLUMN time_log.cycle IS 'Cycle number recorded for this interval.';
COMMENT ON COLUMN time_log.start_time IS 'When this interval started.';
COMMENT ON COLUMN time_log.stop_time IS 'When this interval stopped.';
COMMENT ON COLUMN time_log.interruption_minutes IS 'Minutes of interruption inside this interval.';
COMMENT ON COLUMN time_log.comments IS 'Notes about this interval.';
