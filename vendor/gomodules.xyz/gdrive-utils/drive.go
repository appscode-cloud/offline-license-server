package gdrive_utils

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

func AddPermission(svc *drive.Service, docId string, email string, role string) (string, error) {
	perm, err := svc.Permissions.Create(docId, &drive.Permission{
		EmailAddress: email,
		Role:         role,
		Type:         "user",
	}).Fields("id").Do()
	if err != nil {
		return "", err
	}
	return perm.Id, nil
}

// https://developers.google.com/youtube/v3/getting-started#partial
// parts vs fields
// parts is top level section
// fields are fields inside that section

func RevokePermission(svc *drive.Service, docId string, email string) error {
	call := svc.Permissions.List(docId)
	call = call.Fields("permissions(id,role,type,emailAddress)")
	var perms []*drive.Permission
	err := call.Pages(context.TODO(), func(resp *drive.PermissionList) error {
		perms = append(perms, resp.Permissions...)
		return nil
	})

	for _, perm := range perms {
		if perm.EmailAddress == email {
			return svc.Permissions.Delete(docId, perm.Id).Do()
		}
	}
	return err
}

func FindParentFolderId(svc *drive.Service, configDocId string) (string, error) {
	d, err := svc.Files.Get(configDocId).Fields("parents").Do()
	if err != nil {
		return "", err
	}
	return d.Parents[0], nil
}

func GetFolderId(svc *drive.Service, configDocId string, p string) (string, error) {
	parentFolderId, err := FindParentFolderId(svc, configDocId)
	if err != nil {
		return "", errors.Wrap(err, "failed to detect root folder id")
	}
	p = path.Clean(p)
	// empty path (p == "")
	if p == "." {
		return parentFolderId, nil
	}
	folders := strings.Split(p, "/")
	for _, folderName := range folders {
		// https://developers.google.com/drive/api/v3/search-files
		q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents", folderName, parentFolderId)
		files, err := svc.Files.List().Q(q).Spaces("drive").Do()
		if err != nil {
			return "", errors.Wrapf(err, "failed to find folder %s inside parent folder %s", folderName, parentFolderId)
		}
		if len(files.Files) > 0 {
			parentFolderId = files.Files[0].Id
		} else {
			// https://developers.google.com/drive/api/v3/folder#java
			folderMetadata := &drive.File{
				Name:     folderName,
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{parentFolderId},
			}
			folder, err := svc.Files.Create(folderMetadata).Fields("id").Do()
			if err != nil {
				return "", errors.Wrapf(err, "failed to create folder %s inside parent folder %s", folderName, parentFolderId)
			}
			parentFolderId = folder.Id
		}
	}
	return parentFolderId, nil
}

func CopyDoc(svcDrive *drive.Service, svcDocs *docs.Service, templateDocId string, folderId string, docName string, replacements map[string]string) (string, error) {
	var docId string

	// https://developers.google.com/drive/api/v3/search-files
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.document' and '%s' in parents", docName, folderId)
	files, err := svcDrive.Files.List().Q(q).Spaces("drive").Do()
	if err != nil {
		return "", errors.Wrapf(err, "failed to find doc %s inside parent folder %s", docName, folderId)
	}
	if len(files.Files) > 0 {
		docId = files.Files[0].Id
	} else {
		// https://developers.google.com/docs/api/how-tos/documents#copying_an_existing_document
		copyMetadata := &drive.File{
			Name:    docName,
			Parents: []string{folderId},
		}
		doc, err := svcDrive.Files.Copy(templateDocId, copyMetadata).Fields("id").Do()
		if err != nil {
			return "", errors.Wrapf(err, "failed to copy template doc %s into folder %s", templateDocId, folderId)
		}
		docId = doc.Id

		if len(replacements) > 0 {
			// https://developers.google.com/docs/api/how-tos/merge
			req := &docs.BatchUpdateDocumentRequest{
				Requests: make([]*docs.Request, 0, len(replacements)),
			}
			for k, v := range replacements {
				req.Requests = append(req.Requests, &docs.Request{
					ReplaceAllText: &docs.ReplaceAllTextRequest{
						ContainsText: &docs.SubstringMatchCriteria{
							MatchCase: true,
							Text:      k,
						},
						ReplaceText: v,
					},
				})
			}
			_, err := svcDocs.Documents.BatchUpdate(docId, req).Do()
			if err != nil {
				return "", errors.Wrapf(err, "failed to replace template fields in doc %s", docId)
			}
		}
	}
	return docId, nil
}
