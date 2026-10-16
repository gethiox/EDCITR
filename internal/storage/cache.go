package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gethiox/EDCITR/internal/inara/inapi/events/commander_profile"
)

var (
	subdirRootData       = "inapi"
	subdirProfileInfo    = filepath.Join(subdirRootData, "profiles") //
	subdirPictures       = filepath.Join(subdirRootData, "pictures") // CmdrID/pfp.jpg
	subdirCommanderNames = filepath.Join(subdirRootData, "commanders")
)

const (
	filenameSuffixJson = ".json"
	filenameSuffixJpg  = ".jpg"
)

var (
	HasNoProfileError = errors.New("profile does not exist")
	ProfileNotFound   = errors.New("profile not found")
	HasNoPictureError = errors.New("picture does not exist")
	PictureNotFound   = errors.New("picture not found")
)

type InfoWithTime struct {
	ProfileInfo     commander_profile.CommanderProfileInfo
	ProfileInfoTime time.Time
}

type Cache struct {
	root string

	statsReady      atomic.Bool
	picturesPresent atomic.Int64
	picturesMissing atomic.Int64
	infosPresent    atomic.Int64
	infoMissing     atomic.Int64
}

func GetCache(appName string) (*Cache, error) {
	if appName == "" {
		return nil, errors.New("appName cannot be empty")
	}

	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("get user cache directory failed: %w", err)
	}

	if cacheRoot == "" {
		return nil, errors.New("user cache directory is an empty string")
	}

	appCacheDirectory := filepath.Join(cacheRoot, appName)

	err = os.MkdirAll(appCacheDirectory, 0755)
	if err != nil {
		return nil, fmt.Errorf("create cache directory failed: %w", err)
	}

	cache := Cache{root: appCacheDirectory}

	err = cache.initDirectoryStructure()
	if err != nil {
		return nil, fmt.Errorf("create cache directory failed: %w", err)
	}

	go func() {
		var ip, im, pp, pm int

		pp, pm = cache.getPictureCount()
		ip, im = cache.getInfoCount()

		cache.infosPresent.Store(int64(ip))
		cache.infoMissing.Store(int64(im))
		cache.picturesPresent.Store(int64(pp))
		cache.picturesMissing.Store(int64(pm))

		cache.statsReady.Store(true)
	}()

	return &cache, nil
}

func (c *Cache) Root() string {
	return c.root
}

func (c *Cache) getPictureCount() (present, missing int) {
	prefixEntries, err := os.ReadDir(filepath.Join(c.root, subdirPictures))
	if err != nil {
		return 0, 0
	}

	for _, prefixEntry := range prefixEntries {
		images, err := os.ReadDir(filepath.Join(c.root, subdirPictures, prefixEntry.Name()))
		if err != nil {
			return 0, 0
		}

		for _, img := range images {
			stat, err := os.Stat(filepath.Join(c.root, subdirPictures, prefixEntry.Name(), img.Name()))
			if err != nil {
				return 0, 0
			}

			if stat.Size() == 0 {
				missing++
			} else {
				present++
			}
		}
	}

	return present, missing
}

func (c *Cache) getInfoCount() (present, missing int) {
	prefixEntries, err := os.ReadDir(filepath.Join(c.root, subdirCommanderNames))
	if err != nil {
		return 0, 0
	}

	for _, prefixEntry := range prefixEntries {
		symlinks, err := os.ReadDir(filepath.Join(c.root, subdirCommanderNames, prefixEntry.Name()))
		if err != nil {
			return 0, 0
		}

		for _, symlink := range symlinks {
			data, err := os.ReadFile(filepath.Join(c.root, subdirCommanderNames, prefixEntry.Name(), symlink.Name()))
			if err != nil {
				return 0, 0
			}

			var sl PseudoSymlink
			err = json.Unmarshal(data, &sl)
			if err != nil {
				return 0, 0
			}

			if sl.ProfileExists {
				present++
			} else {
				missing++
			}
		}
	}

	return present, missing
}

func (c *Cache) GetPresentPictureCount() int {
	return int(c.picturesPresent.Load())
}

func (c *Cache) GetMissingPictureCount() int {
	return int(c.picturesMissing.Load())
}

func (c *Cache) GetPresentInfoCount() int {
	return int(c.infosPresent.Load())
}

func (c *Cache) GetMissingInfoCount() int {
	return int(c.infoMissing.Load())
}

type PseudoSymlink struct {
	CmdrName      string `json:"name"`
	CmdrID        uint64 `json:"id"`
	ProfileExists bool   `json:"profile_exists"`
}

func (c *Cache) CacheDirectory() string {
	return c.root
}

func (c *Cache) initDirectoryStructure() error {
	for _, dir := range []string{subdirProfileInfo, subdirPictures, subdirCommanderNames} {
		err := os.MkdirAll(filepath.Join(c.root, dir), 0755)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("failed to create dir: %w", err)
		}
	}
	return nil
}

func (c *Cache) ClearCache() error {
	entries, err := os.ReadDir(c.root)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(c.root, entry.Name())
		err := os.RemoveAll(entryPath)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %w", entryPath, err)
		}
	}

	c.infoMissing.Store(0)
	c.infosPresent.Store(0)
	c.picturesMissing.Store(0)
	c.picturesPresent.Store(0)

	return nil
}

func (c *Cache) SaveMissingProfileInfo(cmdrName string) error {
	if cmdrName == "" {
		return errors.New("cmdrName cannot be empty")
	}

	err := c.saveSymlink(cmdrName, 0, false)
	if err != nil {
		return fmt.Errorf("failed to save symlink: %w", err)
	}

	c.infoMissing.Add(1)

	return nil
}

func (c *Cache) saveSymlink(cmdrName string, cmdrID uint64, profileExists bool) error {
	symlink := PseudoSymlink{
		CmdrName:      cmdrName,
		CmdrID:        cmdrID,
		ProfileExists: profileExists,
	}

	symlinkData, err := json.Marshal(symlink, jsontext.Multiline(false))
	if err != nil {
		return fmt.Errorf("failed to marshal symlink data: %w", err)
	}

	symlinkDataDir, symlinkDataFilename := dirAndFilenameFromCmdrName(cmdrName)
	symlinkDataFilename += filenameSuffixJson
	symlinkDataDirAbsolute := filepath.Join(c.root, subdirCommanderNames, symlinkDataDir)
	symlinkDataFilenameAbsolute := filepath.Join(symlinkDataDirAbsolute, symlinkDataFilename)

	err = os.MkdirAll(symlinkDataDirAbsolute, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(symlinkDataFilenameAbsolute, symlinkData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write symlink data: %w", err)
	}

	return nil
}

func (c *Cache) SaveProfileInfo(cmdrName string, info commander_profile.CommanderProfileInfo) error {
	if cmdrName == "" {
		return errors.New("cmdrName cannot be empty")
	}

	infoData, err := json.Marshal(info, jsontext.Multiline(false))
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	infoDataDir, infoDataFilename := dirAndFilenameFromCmdrID(info.UserID)
	infoDataFilename += filenameSuffixJson
	infoDataDirAbsolute := filepath.Join(c.root, subdirProfileInfo, infoDataDir)
	infoDataFilenameAbsolute := filepath.Join(infoDataDirAbsolute, infoDataFilename)

	err = os.MkdirAll(infoDataDirAbsolute, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(infoDataFilenameAbsolute, infoData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write info data into a file: %w", err)
	}

	err = c.saveSymlink(cmdrName, info.UserID, true)
	if err != nil {
		return fmt.Errorf("failed to save symlink: %w", err)
	}

	c.infosPresent.Add(1)

	return nil
}

func (c *Cache) LoadProfileInfoFromID(cmdrID uint64) (commander_profile.CommanderProfileInfo, time.Time, error) {
	infoDataDir, infoDataFilename := dirAndFilenameFromCmdrID(cmdrID)
	infoDataFilename += filenameSuffixJson
	infoDataDirAbsolute := filepath.Join(c.root, subdirProfileInfo, infoDataDir)
	infoDataFilenameAbsolute := filepath.Join(infoDataDirAbsolute, infoDataFilename)

	infoData, err := os.ReadFile(infoDataFilenameAbsolute)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return commander_profile.CommanderProfileInfo{}, time.Time{}, ProfileNotFound
		}
		return commander_profile.CommanderProfileInfo{}, time.Time{}, fmt.Errorf("failed to read profile info: %w", err)
	}

	stat, err := os.Stat(infoDataFilenameAbsolute)
	if err != nil {
		return commander_profile.CommanderProfileInfo{}, time.Time{}, fmt.Errorf("failed to stat profile info: %w", err)
	}

	var info commander_profile.CommanderProfileInfo
	err = json.Unmarshal(infoData, &info)
	if err != nil {
		return commander_profile.CommanderProfileInfo{}, time.Time{}, fmt.Errorf("failed to unmarshal profile info: %w", err)
	}

	return info, stat.ModTime(), nil
}

func (c *Cache) LoadProfileInfoFromName(cmdrName string) (commander_profile.CommanderProfileInfo, time.Time, error) {
	if cmdrName == "" {
		return commander_profile.CommanderProfileInfo{}, time.Time{}, errors.New("cmdrName cannot be empty")
	}

	symlinkDataDir, symlinkDataFilename := dirAndFilenameFromCmdrName(cmdrName)
	symlinkDataFilename += filenameSuffixJson
	symlinkDataDirAbsolute := filepath.Join(c.root, subdirCommanderNames, symlinkDataDir)
	symlinkDataFilenameAbsolute := filepath.Join(symlinkDataDirAbsolute, symlinkDataFilename)

	symlinkData, err := os.ReadFile(symlinkDataFilenameAbsolute)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return commander_profile.CommanderProfileInfo{}, time.Time{}, ProfileNotFound
		}
		return commander_profile.CommanderProfileInfo{}, time.Time{}, fmt.Errorf("failed to read symlink: %w", err)
	}

	stat, err := os.Stat(symlinkDataFilenameAbsolute)
	if err != nil {
		return commander_profile.CommanderProfileInfo{}, time.Time{}, fmt.Errorf("failed to stat symlink: %w", err)
	}

	var symlink PseudoSymlink
	err = json.Unmarshal(symlinkData, &symlink)
	if err != nil {
		return commander_profile.CommanderProfileInfo{}, time.Time{}, fmt.Errorf("failed to unmarshal symlink: %w", err)
	}

	if !symlink.ProfileExists {
		return commander_profile.CommanderProfileInfo{}, stat.ModTime(), HasNoProfileError
	}

	return c.LoadProfileInfoFromID(symlink.CmdrID)
}

func (c *Cache) SaveProfilePicture(CmdrID uint64, img image.Image) error {
	pictureDataDir, pictureDataFilename := dirAndFilenameFromCmdrID(CmdrID)
	pictureDataFilename += filenameSuffixJpg
	pictureDataDirAbsolute := filepath.Join(c.root, subdirPictures, pictureDataDir)
	pictureDataFilenameAbsolute := filepath.Join(pictureDataDirAbsolute, pictureDataFilename)

	err := os.MkdirAll(pictureDataDirAbsolute, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var pictureData = new(bytes.Buffer)
	err = jpeg.Encode(pictureData, img, nil)
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	err = os.WriteFile(pictureDataFilenameAbsolute, pictureData.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	c.picturesPresent.Add(1)

	return nil
}

func (c *Cache) SaveMissingProfilePicture(CmdrID uint64) error {
	pictureDataDir, pictureDataFilename := dirAndFilenameFromCmdrID(CmdrID)
	pictureDataFilename += filenameSuffixJpg
	pictureDataDirAbsolute := filepath.Join(c.root, subdirPictures, pictureDataDir)
	pictureDataFilenameAbsolute := filepath.Join(pictureDataDirAbsolute, pictureDataFilename)

	err := os.MkdirAll(pictureDataDirAbsolute, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(pictureDataFilenameAbsolute, []byte{}, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	c.picturesMissing.Add(1)

	return nil
}

func (c *Cache) LoadProfilePicture(CmdrID uint64) (image.Image, time.Time, error) {
	pictureDataDir, pictureDataFilename := dirAndFilenameFromCmdrID(CmdrID)
	pictureDataFilename += filenameSuffixJpg
	pictureDataDirAbsolute := filepath.Join(c.root, subdirPictures, pictureDataDir)
	pictureDataFilenameAbsolute := filepath.Join(pictureDataDirAbsolute, pictureDataFilename)

	pictureData, err := os.ReadFile(pictureDataFilenameAbsolute)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, time.Time{}, PictureNotFound
		}
		return nil, time.Time{}, fmt.Errorf("failed to read image file: %w", err)
	}

	stat, err := os.Stat(pictureDataFilenameAbsolute)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to stat image file: %w", err)
	}

	if stat.Size() == 0 {
		return nil, stat.ModTime(), HasNoPictureError
	}

	img, err := jpeg.Decode(bytes.NewReader(pictureData))
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to decode image file: %w", err)
	}

	return img, stat.ModTime(), nil
}

func dirAndFilenameFromCmdrID(cmdrID uint64) (string, string) {
	sha := uint64ToSha256(cmdrID)
	return fmt.Sprintf("%x", sha[:1]), fmt.Sprintf("%x", sha[1:])
}

func dirAndFilenameFromCmdrName(cmdrName string) (string, string) {
	sha := sha256.Sum256([]byte(cmdrName))
	return fmt.Sprintf("%x", sha[:1]), fmt.Sprintf("%x", sha[1:])
}

func uint64ToSha256(i uint64) [32]byte {
	return sha256.Sum256([]byte(fmt.Sprintf("%016s", strconv.FormatUint(i, 16))))
}
