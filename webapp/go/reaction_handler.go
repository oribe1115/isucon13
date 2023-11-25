package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type ReactionModel struct {
	ID           int64  `db:"id"`
	EmojiName    string `db:"emoji_name"`
	UserID       int64  `db:"user_id"`
	LivestreamID int64  `db:"livestream_id"`
	CreatedAt    int64  `db:"created_at"`
}

type Reaction struct {
	ID         int64      `json:"id"`
	EmojiName  string     `json:"emoji_name"`
	User       User       `json:"user"`
	Livestream Livestream `json:"livestream"`
	CreatedAt  int64      `json:"created_at"`
}

type PostReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

func getReactionsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	//tx, err := dbConn.BeginTxx(ctx, nil)
	//if err != nil {
	//	return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	//}
	//defer tx.Rollback()

	query := "SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC"
	if c.QueryParam("limit") != "" {
		limit, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "limit query parameter must be integer")
		}
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	reactionModels := []ReactionModel{}
	if err := dbConn.SelectContext(ctx, &reactionModels, query, livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "failed to get reactions")
	}
	livestream, err := filledLivestreamResponse(ctx, int64(livestreamID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "failed to get filledLivestreamResponse")
	}

	reactions := make([]Reaction, len(reactionModels))
	for i := range reactionModels {
		reaction, err := fillReactionResponseImpl(ctx, reactionModels[i], &livestream)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill reaction: "+err.Error())
		}
		reactions[i] = reaction
	}

	//if err := tx.Commit(); err != nil {
	//	return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	//}

	return cJSON(c, http.StatusOK, reactions)
}

func postReactionHandler(c echo.Context) error {
	ctx := c.Request().Context()
	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	// error already checked
	sess, _ := session.Get(defaultSessionIDKey, c)
	// existence already checked
	userID := sess.Values[defaultUserIDKey].(int64)

	var req *PostReactionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	//tx, err := dbConn.BeginTxx(ctx, nil)
	//if err != nil {
	//	return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	//}
	//defer tx.Rollback()

	reactionModel := ReactionModel{
		UserID:       int64(userID),
		LivestreamID: int64(livestreamID),
		EmojiName:    req.EmojiName,
		CreatedAt:    time.Now().Unix(),
	}

	result, err := dbConn.NamedExecContext(ctx, "INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (:user_id, :livestream_id, :emoji_name, :created_at)", reactionModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert reaction: "+err.Error())
	}

	reactionID, err := result.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get last inserted reaction id: "+err.Error())
	}
	reactionModel.ID = reactionID

	reaction, err := fillReactionResponse(ctx, reactionModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill reaction: "+err.Error())
	}

	//if err := tx.Commit(); err != nil {
	//	return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	//}

	return cJSON(c, http.StatusCreated, reaction)
}

func fillReactionResponse(ctx context.Context, reactionModel ReactionModel) (Reaction, error) {
	return fillReactionResponseImpl(ctx, reactionModel, nil)
}

func fillReactionResponseImpl(ctx context.Context, reactionModel ReactionModel, livestreamDefault *Livestream) (Reaction, error) {
	userModel, err := userCache.Get(ctx, reactionModel.UserID)
	if err != nil {
		return Reaction{}, err
	}
	//userModel := UserModel{}
	//if err := tx.GetContext(ctx, &userModel, "SELECT * FROM users WHERE id = ?", reactionModel.UserID); err != nil {
	//	return Reaction{}, err
	//}
	user, err := fillUserResponse(ctx, *userModel)
	if err != nil {
		return Reaction{}, err
	}

	var livestream Livestream
	if livestreamDefault != nil {
		livestream = *livestreamDefault
	} else {
		livestream, err = filledLivestreamResponse(ctx, reactionModel.LivestreamID)
		if err != nil {
			return Reaction{}, err
		}
	}

	reaction := Reaction{
		ID:         reactionModel.ID,
		EmojiName:  reactionModel.EmojiName,
		User:       user,
		Livestream: livestream,
		CreatedAt:  reactionModel.CreatedAt,
	}

	return reaction, nil
}
