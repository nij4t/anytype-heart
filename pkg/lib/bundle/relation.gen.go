/*
Code generated by pkg/lib/bundle/generator. DO NOT EDIT.
source: pkg/lib/bundle/relations.json
*/
package bundle

import "github.com/anytypeio/go-anytype-middleware/pkg/lib/pb/relation"

type RelationKey string

func (rk RelationKey) String() string {
	return string(rk)
}

const (
	RelationKeyTag                   RelationKey = "tag"
	RelationKeyCamera                RelationKey = "camera"
	RelationKeyHeightInPixels        RelationKey = "heightInPixels"
	RelationKeyCreatedDate           RelationKey = "createdDate"
	RelationKeyToBeDeletedDate       RelationKey = "toBeDeletedDate"
	RelationKeyDone                  RelationKey = "done"
	RelationKeyDateOfBirth           RelationKey = "dateOfBirth"
	RelationKeyThumbnailImage        RelationKey = "thumbnailImage"
	RelationKeyAttachments           RelationKey = "attachments"
	RelationKeyLinkedTasks           RelationKey = "linkedTasks"
	RelationKeyIconImage             RelationKey = "iconImage"
	RelationKeyReleasedYear          RelationKey = "releasedYear"
	RelationKeyCoverScale            RelationKey = "coverScale"
	RelationKeyLinkedProjects        RelationKey = "linkedProjects"
	RelationKeyAudioAlbum            RelationKey = "audioAlbum"
	RelationKeyStatus                RelationKey = "status"
	RelationKeyDurationInSeconds     RelationKey = "durationInSeconds"
	RelationKeyAperture              RelationKey = "aperture"
	RelationKeyLastModifiedDate      RelationKey = "lastModifiedDate"
	RelationKeyRecommendedRelations  RelationKey = "recommendedRelations"
	RelationKeyCreator               RelationKey = "creator"
	RelationKeyRecommendedLayout     RelationKey = "recommendedLayout"
	RelationKeyLastOpenedDate        RelationKey = "lastOpenedDate"
	RelationKeyArtist                RelationKey = "artist"
	RelationKeyDueDate               RelationKey = "dueDate"
	RelationKeyIconEmoji             RelationKey = "iconEmoji"
	RelationKeyCoverType             RelationKey = "coverType"
	RelationKeyCoverY                RelationKey = "coverY"
	RelationKeySizeInBytes           RelationKey = "sizeInBytes"
	RelationKeyCollectionOf          RelationKey = "collectionOf"
	RelationKeyAddedDate             RelationKey = "addedDate"
	RelationKeyAssignee              RelationKey = "assignee"
	RelationKeyExposure              RelationKey = "exposure"
	RelationKeyAudioGenre            RelationKey = "audioGenre"
	RelationKeyName                  RelationKey = "name"
	RelationKeyFocalRatio            RelationKey = "focalRatio"
	RelationKeyPriority              RelationKey = "priority"
	RelationKeyFileMimeType          RelationKey = "fileMimeType"
	RelationKeyType                  RelationKey = "type"
	RelationKeyLayout                RelationKey = "layout"
	RelationKeyAudioAlbumTrackNumber RelationKey = "audioAlbumTrackNumber"
	RelationKeyPlaceOfBirth          RelationKey = "placeOfBirth"
	RelationKeyComposer              RelationKey = "composer"
	RelationKeyCoverX                RelationKey = "coverX"
	RelationKeyDescription           RelationKey = "description"
	RelationKeyId                    RelationKey = "id"
	RelationKeyCameraIso             RelationKey = "cameraIso"
	RelationKeyCoverId               RelationKey = "coverId"
	RelationKeyLastModifiedBy        RelationKey = "lastModifiedBy"
	RelationKeyWidthInPixels         RelationKey = "widthInPixels"
	RelationKeySetOf                 RelationKey = "setOf"
	RelationKeyGender                RelationKey = "gender"
	RelationKeyFileExt               RelationKey = "fileExt"
	RelationKeyFeaturedRelations     RelationKey = "featuredRelations"
)

var (
	relations = map[RelationKey]*relation.Relation{
		RelationKeyAddedDate: {

			DataSource:  relation.Relation_details,
			Description: "Date when the file were added into the anytype",
			Format:      relation.RelationFormat_date,
			Hidden:      false,
			Key:         "addedDate",
			Name:        "Added date",
			ReadOnly:    false,
		},
		RelationKeyAperture: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "aperture",
			Name:        "Camera Aperture",
			ReadOnly:    false,
		},
		RelationKeyArtist: {

			DataSource:  relation.Relation_details,
			Description: "Name of artist",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "artist",
			Name:        "Artist",
			ReadOnly:    false,
		},
		RelationKeyAssignee: {

			DataSource:  relation.Relation_details,
			Description: "Person who is responsible for this task or object",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "assignee",
			Name:        "Assignee",
			ObjectTypes: []string{TypePrefix + "profile"},
			ReadOnly:    false,
		},
		RelationKeyAttachments: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "attachments",
			Name:        "Attachments",
			ObjectTypes: []string{TypePrefix + "file", TypePrefix + "image", TypePrefix + "video", TypePrefix + "audio"},
			ReadOnly:    false,
		},
		RelationKeyAudioAlbum: {

			DataSource:  relation.Relation_details,
			Description: "Audio record's album name",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "audioAlbum",
			Name:        "Album",
			ReadOnly:    false,
		},
		RelationKeyAudioAlbumTrackNumber: {

			DataSource:  relation.Relation_details,
			Description: "Number of the track in the",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "audioAlbumTrackNumber",
			Name:        "Track #",
			ReadOnly:    false,
		},
		RelationKeyAudioGenre: {

			DataSource:  relation.Relation_details,
			Description: "Audio record's genre name",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "audioGenre",
			Name:        "Genre",
			ReadOnly:    false,
		},
		RelationKeyCamera: {

			DataSource:  relation.Relation_details,
			Description: "Camera used to capture image or video",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "camera",
			Name:        "Camera",
			ReadOnly:    false,
		},
		RelationKeyCameraIso: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "cameraIso",
			Name:        "ISO",
			ReadOnly:    false,
		},
		RelationKeyCollectionOf: {

			DataSource:  relation.Relation_details,
			Description: "Point to the object types that can be added to collection. Empty means any object type can be added to the collection",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "collectionOf",
			Name:        "Collection of",
			ObjectTypes: []string{TypePrefix + "objectType"},
			ReadOnly:    false,
		},
		RelationKeyComposer: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "composer",
			Name:        "Composer",
			ReadOnly:    false,
		},
		RelationKeyCoverId: {

			DataSource:  relation.Relation_details,
			Description: "Can contains image hash, color or prebuild bg id, depends on coverType relation",
			Format:      relation.RelationFormat_title,
			Hidden:      true,
			Key:         "coverId",
			Name:        "Cover image or color",
			ReadOnly:    false,
		},
		RelationKeyCoverScale: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_number,
			Hidden:      true,
			Key:         "coverScale",
			Name:        "Cover scale",
			ReadOnly:    false,
		},
		RelationKeyCoverType: {

			DataSource:  relation.Relation_details,
			Description: "1-image, 2-color, 3-gradient, 4-prebuilt bg image. Value stored in coverId",
			Format:      relation.RelationFormat_number,
			Hidden:      true,
			Key:         "coverType",
			Name:        "Cover type",
			ReadOnly:    false,
		},
		RelationKeyCoverX: {

			DataSource:  relation.Relation_details,
			Description: "Image x offset of the provided image",
			Format:      relation.RelationFormat_number,
			Hidden:      true,
			Key:         "coverX",
			Name:        "Cover x offset",
			ReadOnly:    false,
		},
		RelationKeyCoverY: {

			DataSource:  relation.Relation_details,
			Description: "Image y offset of the provided image",
			Format:      relation.RelationFormat_number,
			Hidden:      true,
			Key:         "coverY",
			Name:        "Cover y offset",
			ReadOnly:    false,
		},
		RelationKeyCreatedDate: {

			DataSource:  relation.Relation_derived,
			Description: "Date when the object was initially created",
			Format:      relation.RelationFormat_date,
			Hidden:      false,
			Key:         "createdDate",
			Name:        "Creation date",
			ReadOnly:    true,
		},
		RelationKeyCreator: {

			DataSource:  relation.Relation_derived,
			Description: "Human which created this object",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "creator",
			Name:        "Created by",
			ObjectTypes: []string{TypePrefix + "profile"},
			ReadOnly:    true,
		},
		RelationKeyDateOfBirth: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_date,
			Hidden:      false,
			Key:         "dateOfBirth",
			Name:        "Date of birth",
			ReadOnly:    false,
		},
		RelationKeyDescription: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_description,
			Hidden:      false,
			Key:         "description",
			Name:        "Description",
			ReadOnly:    false,
		},
		RelationKeyDone: {

			DataSource:  relation.Relation_details,
			Description: "Done checkbox used to render action layout. ",
			Format:      relation.RelationFormat_checkbox,
			Hidden:      false,
			Key:         "done",
			Name:        "Done",
			ReadOnly:    false,
		},
		RelationKeyDueDate: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_date,
			Hidden:      false,
			Key:         "dueDate",
			Name:        "Due date",
			ReadOnly:    false,
		},
		RelationKeyDurationInSeconds: {

			DataSource:  relation.Relation_details,
			Description: "Duration of audio/video file in seconds",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "durationInSeconds",
			Name:        "Duration(sec)",
			ReadOnly:    false,
		},
		RelationKeyExposure: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "exposure",
			Name:        "Camera Exposure",
			ReadOnly:    false,
		},
		RelationKeyFeaturedRelations: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_object,
			Hidden:      true,
			Key:         "featuredRelations",
			Name:        "Featured relations management will be \u2028implemented later.",
			ObjectTypes: []string{TypePrefix + "relation"},
			ReadOnly:    false,
		},
		RelationKeyFileExt: {

			DataSource:  relation.Relation_derived,
			Description: "",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "fileExt",
			Name:        "File extension",
			ReadOnly:    false,
		},
		RelationKeyFileMimeType: {

			DataSource:  relation.Relation_details,
			Description: "Mime type of object",
			Format:      relation.RelationFormat_title,
			Hidden:      true,
			Key:         "fileMimeType",
			Name:        "Mime type",
			ReadOnly:    false,
		},
		RelationKeyFocalRatio: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "focalRatio",
			Name:        "Focal ratio",
			ReadOnly:    false,
		},
		RelationKeyGender: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_status,
			Hidden:      false,
			Key:         "gender",
			Name:        "Gender",
			ReadOnly:    false,
		},
		RelationKeyHeightInPixels: {

			DataSource:  relation.Relation_details,
			Description: "Height of image/video in pixels",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "heightInPixels",
			Name:        "Height(px)",
			ReadOnly:    true,
		},
		RelationKeyIconEmoji: {

			DataSource:  relation.Relation_details,
			Description: "1 emoji(can contains multiple UTF symbols) used as an icon",
			Format:      relation.RelationFormat_emoji,
			Hidden:      true,
			Key:         "iconEmoji",
			Name:        "Emoji",
			ReadOnly:    false,
		},
		RelationKeyIconImage: {

			DataSource:  relation.Relation_details,
			Description: "Image icon",
			Format:      relation.RelationFormat_object,
			Hidden:      true,
			Key:         "iconImage",
			Name:        "Image",
			ObjectTypes: []string{TypePrefix + "image"},
			ReadOnly:    false,
		},
		RelationKeyId: {

			DataSource:  relation.Relation_derived,
			Description: "Link to itself. Used in databases",
			Format:      relation.RelationFormat_object,
			Hidden:      true,
			Key:         "id",
			Name:        "Anytype ID",
			ReadOnly:    false,
		},
		RelationKeyLastModifiedBy: {

			DataSource:  relation.Relation_derived,
			Description: "Human which updates the object last time",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "lastModifiedBy",
			Name:        "Last modified by",
			ObjectTypes: []string{TypePrefix + "profile"},
			ReadOnly:    true,
		},
		RelationKeyLastModifiedDate: {

			DataSource:  relation.Relation_derived,
			Description: "Date when the object was modified last time",
			Format:      relation.RelationFormat_date,
			Hidden:      false,
			Key:         "lastModifiedDate",
			Name:        "Last modified date",
			ReadOnly:    true,
		},
		RelationKeyLastOpenedDate: {

			DataSource:  relation.Relation_account,
			Description: "Date when the object was modified last opened",
			Format:      relation.RelationFormat_date,
			Hidden:      false,
			Key:         "lastOpenedDate",
			Name:        "Last opened date",
			ReadOnly:    true,
		},
		RelationKeyLayout: {

			DataSource:  relation.Relation_details,
			Description: "Anytype layout ID(from pb enum)",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "layout",
			Name:        "Layout",
			ReadOnly:    false,
		},
		RelationKeyLinkedProjects: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "linkedProjects",
			Name:        "Linked Projects",
			ObjectTypes: []string{TypePrefix + "project"},
			ReadOnly:    false,
		},
		RelationKeyLinkedTasks: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "linkedTasks",
			Name:        "Linked tasks",
			ObjectTypes: []string{TypePrefix + "task"},
			ReadOnly:    false,
		},
		RelationKeyName: {

			DataSource:  relation.Relation_details,
			Description: "Name of the object",
			Format:      relation.RelationFormat_title,
			Hidden:      false,
			Key:         "name",
			Name:        "Name",
			ReadOnly:    false,
		},
		RelationKeyPlaceOfBirth: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_status,
			Hidden:      false,
			Key:         "placeOfBirth",
			Name:        "Place of birth",
			ReadOnly:    false,
		},
		RelationKeyPriority: {

			DataSource:  relation.Relation_details,
			Description: "Used to order tasks in list/canban",
			Format:      relation.RelationFormat_number,
			Hidden:      true,
			Key:         "priority",
			Name:        "Priority",
			ReadOnly:    false,
		},
		RelationKeyRecommendedLayout: {

			DataSource:  relation.Relation_details,
			Description: "Recommended layout for new templates and objects of specific objec",
			Format:      relation.RelationFormat_number,
			Hidden:      true,
			Key:         "recommendedLayout",
			Name:        "Recommended layout",
			ReadOnly:    false,
		},
		RelationKeyRecommendedRelations: {

			DataSource:  relation.Relation_details,
			Description: "List of recommended relations",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "recommendedRelations",
			Name:        "Recommended relations",
			ObjectTypes: []string{TypePrefix + "relation"},
			ReadOnly:    false,
		},
		RelationKeyReleasedYear: {

			DataSource:  relation.Relation_details,
			Description: "Year when this object were released",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "releasedYear",
			Name:        "Released year",
			ReadOnly:    false,
		},
		RelationKeySetOf: {

			DataSource:  relation.Relation_details,
			Description: "Point to the object types used to aggregate the set. Empty means object of all types will be aggregated ",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "setOf",
			Name:        "Set of",
			ObjectTypes: []string{TypePrefix + "objectType"},
			ReadOnly:    false,
		},
		RelationKeySizeInBytes: {

			DataSource:  relation.Relation_details,
			Description: "Size of file/image in bytes",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "sizeInBytes",
			Name:        "Size(bytes)",
			ReadOnly:    false,
		},
		RelationKeyStatus: {

			DataSource:  relation.Relation_details,
			Description: "Task status?",
			Format:      relation.RelationFormat_status,
			Hidden:      false,
			Key:         "status",
			Name:        "Status",
			ReadOnly:    false,
		},
		RelationKeyTag: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_tag,
			Hidden:      false,
			Key:         "tag",
			Name:        "Tag",
			ReadOnly:    false,
		},
		RelationKeyThumbnailImage: {

			DataSource:  relation.Relation_details,
			Description: "",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "thumbnailImage",
			Name:        "Thumbnail image",
			ObjectTypes: []string{TypePrefix + "image"},
			ReadOnly:    false,
		},
		RelationKeyToBeDeletedDate: {

			DataSource:  relation.Relation_account,
			Description: "Date when the object will be deleted from your device",
			Format:      relation.RelationFormat_date,
			Hidden:      true,
			Key:         "toBeDeletedDate",
			Name:        "Date to delete",
			ReadOnly:    true,
		},
		RelationKeyType: {

			DataSource:  relation.Relation_derived,
			Description: "",
			Format:      relation.RelationFormat_object,
			Hidden:      false,
			Key:         "type",
			Name:        "Object type",
			ObjectTypes: []string{TypePrefix + "objectType"},
			ReadOnly:    false,
		},
		RelationKeyWidthInPixels: {

			DataSource:  relation.Relation_details,
			Description: "Width of image/video in pixels",
			Format:      relation.RelationFormat_number,
			Hidden:      false,
			Key:         "widthInPixels",
			Name:        "Width(px)",
			ReadOnly:    true,
		},
	}
)